// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package collection

import (
	"database/sql"
	"encoding/json"
	"github.com/freetaxii/freetaxii-server/lib/config"
	"github.com/freetaxii/freetaxii-server/lib/headers"
	"github.com/freetaxii/freetaxii-server/lib/services/status"
	"github.com/freetaxii/libtaxii/collection"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
)

type CollectionType struct {
	SysConfig *config.ServerConfigType
}

func (this *CollectionType) CollectionServerHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var taxiiHeader headers.HttpHeaderType
	var statusMsg status.StatusType

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
		statusMessageData := statusMsg.CreateTaxiiStatusMessage("", "BAD_MESSAGE", err.Error())
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
		statusMessageData := statusMsg.CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Can not decode Collection Request")
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: BAD_MESSAGE, can not decode Collection Request")
		}
		w.Write(statusMessageData)
		return
	}

	// Check to make sure their is a message ID in the request message
	if requestMessageData.TaxiiMessage.Id == "" {
		statusMessageData := statusMsg.CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Collection Request message did not include an ID")
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: BAD_MESSAGE, Collection Request message did not include an ID")
		}
		w.Write(statusMessageData)
		return
	}

	if this.SysConfig.Logging.LogLevel >= 1 {
		log.Printf("DEBUG-1: Collection Request from %s with ID: %s", r.RemoteAddr, requestMessageData.TaxiiMessage.Id)
	}

	// Get a list of valid collections for this collection request
	validCollections := this.getValidCollections()

	data := this.createCollectionResponse(requestMessageData.TaxiiMessage.Id, validCollections)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if this.SysConfig.Logging.LogLevel >= 1 {
		log.Println("DEBUG-1: Sending Collection Response to", r.RemoteAddr)
	}
	w.Write(data)

}

// --------------------------------------------------
// Create a TAXII Collection Response Message
// --------------------------------------------------

func (this *CollectionType) createCollectionResponse(inResponseToID string, validCollections map[string]string) []byte {
	tm := collection.NewResponse()
	tm.AddInResponseTo(inResponseToID)

	for k, v := range validCollections {
		c := collection.CreateCollection()
		c.AddName(k)
		c.SetAvailable()
		c.AddDescription(v)
		c.AddVolume(1)
		//c.SetPushMethodToHttpJson()
		c.SetPollServiceToHttpJson("http://test.freetaxii.com:8000/services/poll/")
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

// --------------------------------------------------
// Get list of valid collections
// --------------------------------------------------

func (this *CollectionType) getValidCollections() map[string]string {

	// TODO Read in from a database the collections we offer for this authenticated
	// user and put them in a map
	// TODO switch from a map to a struct so we can track more than just name and description

	// Open connection to database
	filename := this.SysConfig.System.DbFileFullPath
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatalf("Unable to open file %s due to error %v", filename, err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT collection, description FROM Collections")
	if err != nil {
		log.Printf("error running query, %v", err)
	}
	defer rows.Close()

	c := make(map[string]string)

	for rows.Next() {
		var collection string
		var description string
		err = rows.Scan(&collection, &description)

		if err != nil {
			log.Printf("error reading from database, %v", err)
		}

		c[collection] = description
	}

	// c["ip-watch-list"] = "List of interesting IP addresses"
	// c["url-watch-list"] = "List of interesting URL addresses"
	return c
}
