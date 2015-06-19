// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package taxiiserver

import (
	"encoding/json"
	"github.com/freestix/libstix/stix"
	"github.com/freetaxii/freetaxii-server/lib/headers"
	"github.com/freetaxii/libtaxii/pollMessage"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func (this *ServerType) PollServerHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var taxiiHeader headers.HttpHeaderType

	// Log notice of incoming TAXII message
	if this.SysConfig.Logging.LogLevel >= 3 {
		log.Printf("DEBUG-3: Found Message on Poll Server Handler from %s", r.RemoteAddr)
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
	var incomingMessageData pollMessage.PollRequestMessageType
	err = decoder.Decode(&incomingMessageData)

	if err != nil {
		statusMessageData := this.CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Can not decode Poll Request")
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: BAD_MESSAGE, can not decode Poll Request")
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(statusMessageData)
		return
	}

	// Check to make sure there is a message ID in the request message
	if incomingMessageData.Id == "" {
		statusMessageData := this.CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Poll Request message did not include an ID")
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: BAD_MESSAGE, Poll Request message did not include an ID")
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(statusMessageData)
		return
	}

	// Log notice of incomming Poll Request
	if this.SysConfig.Logging.LogLevel >= 1 {
		log.Printf("DEBUG-1: Poll Request from %s for %s with ID: %s", r.RemoteAddr, incomingMessageData.CollectionName, incomingMessageData.Id)
	}

	// --------------------------------------------------
	// Check for valid collection
	// --------------------------------------------------

	currentlyValidCollections := this.SysConfig.GetValidCollections()

	// TODO First check to make sure the value the requested is something they can actually get by their username / subscription / avaliable
	// Based on the collection they are requesting, create a response that contains just the values for that collection

	if _, ok := currentlyValidCollections[incomingMessageData.CollectionName]; ok {
		data := this.createPollResponse(incomingMessageData.Id, incomingMessageData.CollectionName)

		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: Sending Poll Response to", r.RemoteAddr)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(data)
	} else {
		errmsg := "The requested collection \"" + incomingMessageData.CollectionName + "\" does not exist"
		statusMessageData := this.CreateTaxiiStatusMessage("", "DESTINATION_COLLECTION_ERROR", errmsg)
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: DESTINATION_COLLECTION_ERROR, Poll Request asked for a collection that does not exist")
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(statusMessageData)
	}

}

// --------------------------------------------------
// Create a TAXII Poll Response Message
// --------------------------------------------------

func (this *ServerType) createPollResponse(responseid, collectionName string) []byte {
	tm := pollMessage.NewResponse()
	tm.AddInResponseTo(responseid)
	tm.AddCollectionName(collectionName)
	tm.AddResultId("freetaxii-test-service-1")
	tm.AddMessage("This is a test service for FreeTAXII")
	content := tm.NewContentBlock()
	content.SetContentEncodingToJson()
	indicators := this.createIndicatorsJSON(collectionName)
	content.AddContent(indicators)

	data, err := json.Marshal(tm)
	if err != nil {
		// If we can not create a status message then there is something
		// wrong with the APIs and nothing is going to work.
		log.Fatal("Unable to create Poll Response Message")
	}
	return data
}

func (this *ServerType) createIndicatorsJSON(collectionName string) string {

	// Need to pass in the collection name they have requested
	// then go to the database and the get the fields that are needed
	// to populate the correct STIX message.
	// I need a table in the database add the source data and other things to the collections table
	// Create a new table for holding the indicators / observables.
	s := stix.New()
	i1 := s.NewIndicator()
	i1.SetTimestampToNow()

	if collectionName == "ip-watch-list" || collectionName == "url-watch-list" {
		list := []string{
			"176.119.3.108",
			"178.207.85.119",
			"178.63.174.153",
			"188.241.140.212",
			"14.138.73.47",
			"131.72.138.45",
			"62.84.51.39",
			"62.109.23.246",
			"5.101.113.169",
			"213.231.8.30",
			"208.43.25.52",
			"112.208.6.209",
			"115.239.248.87",
			"117.216.190.71",
			"131.72.139.233",
			"129.194.97.21",
			"162.244.35.229",
			"178.219.10.23",
			"184.154.124.203",
			"184.154.146.100",
			"184.154.146.101",
		}
		i1.AddTitle("Malicious IP Addresses")
		i1.AddType("IP Watchlist")
		observable_i1 := i1.NewObservable()
		properties_1 := observable_i1.GetObjectProperties()

		properties_1.AddType("IP Address")

		for _, value := range list {
			properties_1.AddEqualsUriValue(value)
		}

	} else if collectionName == "et-compromised-ips" {

		source1 := stix.CreateInformationSource()
		source1.AddDescriptionText("The Test.FreeTAXII.com Server")
		source1.SetProducedTimeToNow()
		source1.AddReference("http://test.freetaxii.com")

		identity1 := stix.CreateIdentity()
		identity1.AddName("FreeTAXII")
		source1.AddIdentity(identity1)

		contribSource1 := stix.CreateInformationSource()
		identity2 := stix.CreateIdentity()
		identity2.AddName("Emerging Threats Compromised IPs")
		contribSource1.AddIdentity(identity2)
		contribSource1.AddReference("http://rules.emergingthreats.net/blockrules/compromised-ips.txt")

		source1.AddContributingSource(contribSource1)
		i1.AddProducer(source1)

		resp, _ := http.Get("http://rules.emergingthreats.net/blockrules/compromised-ips.txt")
		defer resp.Body.Close()
		rawhtmlbody, _ := ioutil.ReadAll(resp.Body)

		s := string(rawhtmlbody)
		s = strings.TrimSpace(s)
		body := strings.Split(s, "\n")

		i1.AddTitle("Compromised IP Addresses")
		i1.AddType("IP Watchlist")
		observable_i1 := i1.NewObservable()
		properties_1 := observable_i1.GetObjectProperties()

		properties_1.AddType("IP Address")

		for _, value := range body {
			properties_1.AddEqualsUriValue(value)
		}
	}

	var data []byte
	// if this.SysConfig.Poll.output == true {
	data, _ = json.MarshalIndent(s, "", "    ")
	// } else {
	// 	data, _ = json.Marshal(s)
	// }

	return string(data)
}
