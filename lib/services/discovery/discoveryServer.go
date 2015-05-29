// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package discovery

import (
	"encoding/json"
	"github.com/freetaxii/freetaxii-server/lib/headers"
	"github.com/freetaxii/freetaxii-server/lib/services/status"
	"github.com/freetaxii/libtaxii/discovery"
	"log"
	"net/http"
)

type DiscoveryType struct {
	DebugLevel int
}

func (this *DiscoveryType) DiscoveryServerHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var taxiiHeader headers.HttpHeaderType
	var statusMsg status.StatusType

	if this.DebugLevel >= 3 {
		log.Printf("Found Message on Discovery Server Handler from %s", r.RemoteAddr)
	}

	// We need to put this first so that during debugging we can see problems
	// that will generate errors below.
	if this.DebugLevel >= 5 {
		taxiiHeader.DebugHttpRequest(r)
	}

	// --------------------------------------------------
	// Check HTTP Headers for correct TAXII values
	// --------------------------------------------------
	// Send a Status Message on error

	err = taxiiHeader.VerifyHttpTaxiiHeaderValues(r)
	if err != nil {
		if this.DebugLevel >= 2 {
			log.Print(err)
		}

		// If the headers are not right we will not attempt to read the message.
		// This also means that we will not have an InReponseTo ID for the
		// createTaxiiStatusMessage function
		statusMessageData := statusMsg.CreateTaxiiStatusMessage("", "BAD_MESSAGE", err.Error())
		w.Write(statusMessageData)
		return
	}

	// --------------------------------------------------
	// Decode incoming request message
	// --------------------------------------------------
	// Use decoder instead of unmarshal so we can handle stream data
	decoder := json.NewDecoder(r.Body)
	var requestMessageData discovery.TaxiiDiscoveryRequestType
	err = decoder.Decode(&requestMessageData)

	if err != nil {
		statusMessageData := statusMsg.CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Can not decode Discovery Request")
		w.Write(statusMessageData)
		return
	}

	// Check to make sure their is a message ID in the request message
	if requestMessageData.TaxiiMessage.Id == "" {
		statusMessageData := statusMsg.CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Discovery Request message did not include an ID")
		w.Write(statusMessageData)
		return
	}

	if this.DebugLevel >= 1 {
		log.Printf("Discovery Request from %s with ID: %s", r.RemoteAddr, requestMessageData.TaxiiMessage.Id)
	}

	data := this.createDiscoveryResponse(requestMessageData.TaxiiMessage.Id)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)

}

// --------------------------------------------------
// Create a TAXII Discovery Response Message
// --------------------------------------------------

func (this *DiscoveryType) createDiscoveryResponse(responseid string) []byte {
	tm := discovery.NewResponse()
	tm.AddInResponseTo(responseid)

	var s1 discovery.ServiceType
	s1.SetTypeDiscovery()
	s1.SetAvailable()
	s1.SetStandardTaxiiHttpJson()
	s1.AddAddress("http://taxiitest.freetaxii.com/services/discovery")

	var s2 discovery.ServiceType
	s2.SetTypeCollection()
	s2.SetAvailable()
	s2.SetStandardTaxiiHttpJson()
	s2.AddAddress("http://taxiitest.freetaxii.com/services/collection")

	tm.AddService(s1)
	tm.AddService(s2)

	data, err := json.Marshal(tm)
	if err != nil {
		// If we can not create a status message then there is something
		// wrong with the APIs and nothing is going to work.
		log.Fatal("Unable to create Discovery Response Message")
	}
	return data
}
