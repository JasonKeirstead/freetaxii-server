// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package discovery

import (
	"database/sql"
	"encoding/json"
	"github.com/freetaxii/freetaxii-server/lib/config"
	"github.com/freetaxii/freetaxii-server/lib/headers"
	"github.com/freetaxii/freetaxii-server/lib/services/status"
	"github.com/freetaxii/libtaxii/discovery"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
)

type DiscoveryType struct {
	SysConfig      *config.ServerConfigType
	ReloadServices bool
	DiscoveryServicesType
}

type DiscoveryServicesType struct {
	DiscoveryServices []DiscoveryServiceType
}

type DiscoveryServiceType struct {
	ServiceType string
	Available   bool
	Address     string
}

func (this *DiscoveryType) DiscoveryServerHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var taxiiHeader headers.HttpHeaderType
	var statusMsg status.StatusType

	if this.SysConfig.Logging.LogLevel >= 3 {
		log.Printf("DEBUG-3: Found Message on Discovery Server Handler from %s", r.RemoteAddr)
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
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: BAD_MESSAGE", err.Error())
		}
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
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: BAD_MESSAGE, can not decode Discovery Request")
		}
		w.Write(statusMessageData)
		return
	}

	// Check to make sure their is a message ID in the request message
	if requestMessageData.TaxiiMessage.Id == "" {
		statusMessageData := statusMsg.CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Discovery Request message did not include an ID")
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: BAD_MESSAGE, Discovery Request message did not include an ID")
		}
		w.Write(statusMessageData)
		return
	}

	if this.SysConfig.Logging.LogLevel >= 1 {
		log.Printf("DEBUG-1: Discovery Request from %s with ID: %s", r.RemoteAddr, requestMessageData.TaxiiMessage.Id)
	}

	if this.SysConfig.Logging.LogLevel >= 3 {
		log.Println("DEBUG-3: Reload Services is currently set to", this.ReloadServices)
	}
	if this.ReloadServices == true {
		this.loadServices()
		this.ReloadServices = false
		if this.SysConfig.Logging.LogLevel >= 3 {
			log.Println("DEBUG-3: Setting Reload Services to false")
		}
	}

	data := this.createDiscoveryResponse(requestMessageData.TaxiiMessage.Id, this.DiscoveryServices)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
	if this.SysConfig.Logging.LogLevel >= 1 {
		log.Println("DEBUG-1: Sending Discovery Response to", r.RemoteAddr)
	}
}

// --------------------------------------------------
// Create a TAXII Discovery Response Message
// --------------------------------------------------

func (this *DiscoveryType) createDiscoveryResponse(responseid string, ds []DiscoveryServiceType) []byte {
	tm := discovery.NewResponse()
	tm.AddInResponseTo(responseid)

	for _, value := range ds {
		var s discovery.ServiceType

		switch value.ServiceType {
		case "Discovery":
			s.SetTypeDiscovery()
		case "Collection":
			s.SetTypeCollection()
		case "Poll":
			s.SetTypePoll()
		case "Inbox":
			s.SetTypeInbox()
		}

		switch value.Available {
		case true:
			s.SetAvailable()
		case false:
			s.SetUnavailable()
		}
		s.SetStandardTaxiiHttpJson()
		s.AddAddress(value.Address)
		tm.AddService(s)
	}

	data, err := json.Marshal(tm)
	if err != nil {
		// If we can not create a status message then there is something
		// wrong with the APIs and nothing is going to work.
		log.Fatal("Unable to create Discovery Response Message")
	}
	return data
}

// --------------------------------------------------
// Load Services from Database
// --------------------------------------------------
func (this *DiscoveryType) loadServices() {

	if this.SysConfig.Logging.LogLevel >= 1 {
		log.Println("DEBUG-1: Reloading Discovery Services")
	}

	// Clear out existing data so when we reload we do not have a contaminated array
	this.DiscoveryServices = nil

	// Open connection to database
	filename := this.SysConfig.System.DbFileFullPath
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatalf("Unable to open file %s due to error %v", filename, err)
	}
	defer db.Close()

	// Read in services for the discovery server.
	sqlstmt := `SELECT type, available, address 
				FROM Services AS s 
				INNER JOIN ServiceType AS t 
				ON s.typeid = t.id`
	rows, err := db.Query(sqlstmt)
	if err != nil {
		log.Printf("error running query, %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var servicetype string
		var available int
		var address string
		err = rows.Scan(&servicetype, &available, &address)

		if err != nil {
			log.Printf("error reading from database, %v", err)
		}

		var services DiscoveryServiceType
		services.ServiceType = servicetype
		if available == 1 {
			services.Available = true
		} else {
			services.Available = false
		}
		services.Address = address

		// Add services to object
		this.DiscoveryServices = append(this.DiscoveryServices, services)
	}
}
