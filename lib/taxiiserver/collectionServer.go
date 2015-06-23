// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package taxiiserver

import (
	"encoding/json"
	"github.com/freetaxii/freetaxii-server/lib/headers"
	"github.com/freetaxii/libtaxii/messages/collectionMessage"
	"log"
	"net/http"
)

func (this *ServerType) CollectionServerHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var taxiiHeader headers.HttpHeaderType

	if this.SysConfig.Logging.LogLevel >= 3 {
		log.Printf("DEBUG-3: Found Message on Collection Server Handler from %s", r.RemoteAddr)
	}

	// We need to put this first so that during debugging we can see problems
	// that will generate errors below.
	if this.SysConfig.Logging.LogLevel >= 5 {
		taxiiHeader.DebugHttpRequest(r)
	}

	// --------------------------------------------------
	// Check HTTP Headers for correct TAXII values
	// --------------------------------------------------
	// Send a Status Message on error

	err = taxiiHeader.VerifyHttpTaxiiHeaderValues(r)
	if err != nil {
		if this.SysConfig.Logging.LogLevel >= 3 {
			log.Print(err)
		}

		// If the headers are not right we will not attempt to read the message.
		// This also means that we will not have an InReponseTo ID for the
		// createTaxiiStatusMessage function
		statusMessageData := this.CreateTaxiiStatusMessage("", "BAD_MESSAGE", err.Error())
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: BAD_MESSAGE", err.Error())
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(statusMessageData)
		return
	}

	// --------------------------------------------------
	// Decode incoming request message
	// --------------------------------------------------
	// Use decoder instead of unmarshal so we can handle stream data
	decoder := json.NewDecoder(r.Body)
	var incomingMessageData collectionMessage.CollectionRequestMessageType
	err = decoder.Decode(&incomingMessageData)

	if err != nil {
		statusMessageData := this.CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Can not decode Collection Request")
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: BAD_MESSAGE, can not decode Collection Request")
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(statusMessageData)
		return
	}

	// Check to make sure their is a message ID in the request message
	if incomingMessageData.Id == "" {
		statusMessageData := this.CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Collection Request message did not include an ID")
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: BAD_MESSAGE, Collection Request message did not include an ID")
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(statusMessageData)
		return
	}

	if this.SysConfig.Logging.LogLevel >= 1 {
		log.Printf("DEBUG-1: Collection Request from %s with ID: %s", r.RemoteAddr, incomingMessageData.Id)
	}

	// Get a list of valid collections for this collection request
	validCollections := this.SysConfig.GetValidCollections()

	data := this.createCollectionResponse(incomingMessageData.Id, validCollections)
	if this.SysConfig.Logging.LogLevel >= 1 {
		log.Println("DEBUG-1: Sending Collection Response to", r.RemoteAddr)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

// --------------------------------------------------
// Create a TAXII Collection Response Message
// --------------------------------------------------

func (this *ServerType) createCollectionResponse(inResponseToID string, validCollections map[string]string) []byte {
	tm := collectionMessage.NewResponse()
	tm.AddInResponseTo(inResponseToID)

	for k, v := range validCollections {
		c := tm.NewCollection()
		c.AddName(k)
		c.SetAvailable()
		c.AddDescription(v)
		c.AddVolume(1)
		//c.SetPushMethodToHttpJson()
		c.SetPollServiceToHttpJson("http://test.freetaxii.com:8000/services/poll/")
		//c.SetSubscriptionServiceToHttpJson("http://taxiitest.freetaxii.com/services/collection-management/")
		//c.SetInboxServiceToHttpJson("http://taxiitest.freetaxii.com/services/inbox/")
	}

	data, err := json.Marshal(tm)
	if err != nil {
		// If we can not create a status message then there is something
		// wrong with the APIs and nothing is going to work.
		log.Fatal("Unable to create Collection Response Message")
	}
	return data
}
