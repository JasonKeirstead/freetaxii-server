// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package httpHandlers

import (
	"encoding/json"
	"github.com/freetaxii/libtaxii/collection"
	"log"
	"net/http"
)

func (this *HttpHandlersType) CollectionServerHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	if this.DebugLevel >= 2 {
		log.Printf("Found Message on Collection Server Handler from %s", r.RemoteAddr)
	}

	// We need to put this first so that during debugging we can see problems
	// that will generate errors below.
	if this.DebugLevel >= 5 {
		this.debugHttpRequest(r)
	}

	// --------------------------------------------------
	// Check HTTP Headers for correct TAXII values
	// --------------------------------------------------
	// Send a Status Message on error

	err = this.verifyHttpTaxiiHeaderValues(r)
	if err != nil {
		if this.DebugLevel >= 3 {
			log.Print(err)
		}

		// If the headers are not right we will not attempt to read the message.
		// This also means that we will not have an InReponseTo ID for the
		// createTaxiiStatusMessage function
		statusMessageData := CreateTaxiiStatusMessage("", "BAD_MESSAGE", err.Error())
		w.Write(statusMessageData)
		return
	}

	// --------------------------------------------------
	// Decode incoming request message
	// --------------------------------------------------
	// Use decoder instead of unmarshal so we can handle stream data
	decoder := json.NewDecoder(r.Body)
	var requestMessageData collection.TaxiiCollectionRequestType
	err = decoder.Decode(&requestMessageData)

	if err != nil {
		statusMessageData := CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Can not decode Collection Request")
		w.Write(statusMessageData)
		return
	}

	// Check to make sure their is a message ID in the request message
	if requestMessageData.TaxiiMessage.Id == "" {
		statusMessageData := CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Collection Request message did not include an ID")
		w.Write(statusMessageData)
		return
	}

	if this.DebugLevel >= 1 {
		log.Printf("Found TAXII Collection Request Message from %s with ID: %s", r.RemoteAddr, requestMessageData.TaxiiMessage.Id)
	}

	data := CreateCollectionResponse(requestMessageData.TaxiiMessage.Id)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

// --------------------------------------------------
// Create a TAXII Collection Response Message
// --------------------------------------------------

func CreateCollectionResponse(responseid string) []byte {
	tm := collection.NewResponse()
	tm.AddInResponseTo(responseid)

	c1 := collection.CreateCollection()
	c1.AddName("ip-watch-list")
	c1.SetAvailable()
	c1.AddDescription("Data feed of interesting IP addresses")
	c1.AddVolume(1)
	//c1.SetPushMethodToHttpJson()
	c1.SetPollServiceToHttpJson("http://taxiitest.freetaxii.com/services/poll/")
	//c1.SetSubscriptionServiceToHttpJson("http://taxiitest.freetaxii.com/services/collection-management/")
	//c1.SetInboxServiceToHttpJson("http://taxiitest.freetaxii.com/services/inbox/")

	tm.AddCollection(c1)

	data, err := json.Marshal(tm)
	if err != nil {
		// If we can not create a status message then there is something
		// wrong with the APIs and nothing is going to work.
		log.Fatal("Unable to create Collection Response Message")
	}
	return data
}
