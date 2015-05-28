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

	// Get a list of valid collections for this collection request
	validCollections := getValidCollections()

	data := createCollectionResponse(requestMessageData.TaxiiMessage.Id, validCollections)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)

}

// --------------------------------------------------
// Get list of valid collections
// --------------------------------------------------

// TODO Read in from a database the collections we offer for this authenticated
// user and put them in a map

// The key is the collection name and the value is the description
func getValidCollections() map[string]string {
	c := make(map[string]string)
	c["ip-watch-list"] = "List of interesting IP addresses"
	c["url-watch-list"] = "List of interesting URL addresses"
	return c
}

// --------------------------------------------------
// Create a TAXII Collection Response Message
// --------------------------------------------------

func createCollectionResponse(inResponseToID string, validCollections map[string]string) []byte {
	tm := collection.NewResponse()
	tm.AddInResponseTo(inResponseToID)

	for k, v := range validCollections {
		c := collection.CreateCollection()
		c.AddName(k)
		c.SetAvailable()
		c.AddDescription(v)
		c.AddVolume(1)
		//c.SetPushMethodToHttpJson()
		c.SetPollServiceToHttpJson("http://taxiitest.freetaxii.com/services/poll/")
		//c.SetSubscriptionServiceToHttpJson("http://taxiitest.freetaxii.com/services/collection-management/")
		//c.SetInboxServiceToHttpJson("http://taxiitest.freetaxii.com/services/inbox/")

		tm.AddCollection(c)
	}

	data, err := json.Marshal(tm)
	if err != nil {
		// If we can not create a status message then there is something
		// wrong with the APIs and nothing is going to work.
		log.Fatal("Unable to create Collection Response Message")
	}
	return data
}
