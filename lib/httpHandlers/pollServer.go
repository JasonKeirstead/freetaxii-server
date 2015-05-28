// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package httpHandlers

import (
	"encoding/json"
	"github.com/freetaxii/libtaxii/poll"
	"log"
	"net/http"
	"strings"
)

func (this *HttpHandlersType) PollServerHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	if this.DebugLevel >= 2 {
		log.Printf("Found Message on Poll Server Handler from %s", r.RemoteAddr)
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
	var requestMessageData poll.TaxiiPollRequestType
	err = decoder.Decode(&requestMessageData)

	if err != nil {
		statusMessageData := CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Can not decode Poll Request")
		w.Write(statusMessageData)
		return
	}

	// Check to make sure their is a message ID in the request message
	if requestMessageData.TaxiiMessage.Id == "" {
		statusMessageData := CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Poll Request message did not include an ID")
		w.Write(statusMessageData)
		return
	}

	if this.DebugLevel >= 1 {
		log.Printf("Found TAXII Poll Request Message from %s with ID: %s", r.RemoteAddr, requestMessageData.TaxiiMessage.Id)
	}

	// --------------------------------------------------
	// Check for valid collection
	// --------------------------------------------------

	if strings.ToLower(requestMessageData.TaxiiMessage.CollectionName) == "watch-list" {
		data := CreatePollResponse(requestMessageData.TaxiiMessage.Id, "watch-list")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(data)
	} else {
		statusMessageData := CreateTaxiiStatusMessage("", "DESTINATION_COLLECTION_ERROR", "The requested collection does not exist")
		w.Write(statusMessageData)
	}

}

// --------------------------------------------------
// Create a TAXII Discovery Response Message
// --------------------------------------------------

func CreatePollResponse(responseid, collectionName string) []byte {
	tm := poll.NewResponse()
	tm.AddInResponseTo(responseid)
	tm.AddCollectionName(collectionName)

	data, err := json.Marshal(tm)
	if err != nil {
		// If we can not create a status message then there is something
		// wrong with the APIs and nothing is going to work.
		log.Fatal("Unable to create Poll Response Message")
	}
	return data
}
