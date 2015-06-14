// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package poll

import (
	"encoding/json"
	"github.com/freestix/libstix/stix"
	"github.com/freetaxii/freetaxii-server/lib/config"
	"github.com/freetaxii/freetaxii-server/lib/headers"
	"github.com/freetaxii/freetaxii-server/lib/services/status"
	"github.com/freetaxii/libtaxii/poll"
	"log"
	"net/http"
)

type PollType struct {
	SysConfig *config.ServerConfigType
}

func (this *PollType) PollServerHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var taxiiHeader headers.HttpHeaderType
	var statusMsg status.StatusType

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
	var requestMessageData poll.TaxiiPollRequestType
	err = decoder.Decode(&requestMessageData)

	if err != nil {
		statusMessageData := statusMsg.CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Can not decode Poll Request")
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: BAD_MESSAGE, can not decode Poll Request")
		}
		w.Write(statusMessageData)
		return
	}

	// Check to make sure there is a message ID in the request message
	if requestMessageData.TaxiiMessage.Id == "" {
		statusMessageData := statusMsg.CreateTaxiiStatusMessage("", "BAD_MESSAGE", "Poll Request message did not include an ID")
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: BAD_MESSAGE, Poll Request message did not include an ID")
		}
		w.Write(statusMessageData)
		return
	}

	// Log notice of incomming Poll Request
	if this.SysConfig.Logging.LogLevel >= 1 {
		log.Printf("DEBUG-1: Poll Request from %s for %s with ID: %s", r.RemoteAddr, requestMessageData.TaxiiMessage.CollectionName, requestMessageData.TaxiiMessage.Id)
	}

	// --------------------------------------------------
	// Check for valid collection
	// --------------------------------------------------

	currentlyValidCollections := this.SysConfig.GetValidCollections()

	// First check to make sure the value the requested is something they can actually get by their username / subscription / avaliable
	// Based on the collection they are requesting, create a response that contains just the values for that collection
	// Need to pull the values from the database.

	if val, ok := currentlyValidCollections[requestMessageData.TaxiiMessage.CollectionName]; ok {
		data := this.createPollResponse(requestMessageData.TaxiiMessage.Id, val)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: Sending Poll Response to", r.RemoteAddr)
		}
		w.Write(data)
	} else {
		errmsg := "The requested collection \"" + requestMessageData.TaxiiMessage.CollectionName + "\" does not exist"
		statusMessageData := statusMsg.CreateTaxiiStatusMessage("", "DESTINATION_COLLECTION_ERROR", errmsg)
		if this.SysConfig.Logging.LogLevel >= 1 {
			log.Println("DEBUG-1: DESTINATION_COLLECTION_ERROR, Poll Request asked for a collection that does not exist")
		}
		w.Write(statusMessageData)
	}

}

// --------------------------------------------------
// Create a TAXII Discovery Response Message
// --------------------------------------------------

func (this *PollType) createPollResponse(responseid, collectionName string) []byte {
	tm := poll.NewResponse()
	tm.AddInResponseTo(responseid)
	tm.AddCollectionName(collectionName)
	tm.AddResultId("freetaxii-test-service-1")
	tm.AddMessage("This is a test service for FreeTAXII")
	content := poll.CreateContentBlock()
	content.SetContentEncodingToXml()
	indicators := this.createIndicatorsJSON()
	content.AddContent(indicators)
	tm.AddContentBlock(content)

	data, err := json.Marshal(tm)
	if err != nil {
		// If we can not create a status message then there is something
		// wrong with the APIs and nothing is going to work.
		log.Fatal("Unable to create Poll Response Message")
	}
	return data
}

func (this *PollType) createIndicatorsJSON() string {
	s := stix.New()
	i1 := s.NewIndicator()

	i1.SetTimestampToNow()
	i1.AddTitle("Attack 2015-02")
	i1.AddType("IP Watchlist")
	observable_i1 := i1.NewObservable()
	properties_1 := observable_i1.GetObjectProperties()

	properties_1.AddType("URL")
	properties_1.AddEqualsUriValue("http://foo.com")
	properties_1.AddEqualsUriValue("http://bar.com")
	properties_1.AddEqualsUriValue("http://fooandbar.com")

	var data []byte
	data, _ = json.Marshal(s)
	return string(data)
}

func (this *PollType) createIndicatorsXML() string {
	var rawxmldata = `<stix:STIX_Package xsi:schemaLocation="http://cybox.mitre.org/common-2 http://cybox.mitre.org/XMLSchema/common/2.1/cybox_common.xsd  http://cybox.mitre.org/cybox-2 http://cybox.mitre.org/XMLSchema/core/2.1/cybox_core.xsd  http://cybox.mitre.org/default_vocabularies-2 http://cybox.mitre.org/XMLSchema/default_vocabularies/2.1/cybox_default_vocabularies.xsd  http://cybox.mitre.org/objects#URIObject-2 http://cybox.mitre.org/XMLSchema/objects/URI/2.1/URI_Object.xsd  http://stix.mitre.org/Indicator-2 http://stix.mitre.org/XMLSchema/indicator/2.2/indicator.xsd  http://stix.mitre.org/common-1 http://stix.mitre.org/XMLSchema/common/1.2/stix_common.xsd  http://stix.mitre.org/default_vocabularies-1 http://stix.mitre.org/XMLSchema/default_vocabularies/1.2.0/stix_default_vocabularies.xsd  http://stix.mitre.org/stix-1 http://stix.mitre.org/XMLSchema/core/1.2/stix_core.xsd" 
	id="example:Package-8fab937e-b694-11e3-b71c-0800271e87d2" version="1.2">
	<stix:Indicators>
	  <stix:Indicator id="example:Indicator-d81f86b9-975b-bc0b-775e-810c5ad1111" xsi:type="indicator:IndicatorType">
	    <indicator:Title>Malicious IP Addresses</indicator:Title>
	    <indicator:Type xsi:type="stixVocabs:IndicatorTypeVocab-1.0">IP Watchlist</indicator:Type>
	    <indicator:Observable>
	    <cybox:Object><cybox:Properties xsi:type="URIObj:URIObjectType" type="URL">
	    <URIObj:Value condition="Equals">176.119.3.108</URIObj:Value>
	  </cybox:Properties></cybox:Object></indicator:Observable></stix:Indicator>
	  <stix:Indicator id="example:Indicator-d81f86b9-975b-bc0b-775e-810c5ad1111" xsi:type="indicator:IndicatorType">
	    <indicator:Title>Malicious IP Addresses</indicator:Title>
	    <indicator:Type xsi:type="stixVocabs:IndicatorTypeVocab-1.0">IP Watchlist</indicator:Type>
	    <indicator:Observable>
	    <cybox:Object><cybox:Properties xsi:type="URIObj:URIObjectType" type="URL">
	    <URIObj:Value condition="Equals">178.207.85.119</URIObj:Value>
	  </cybox:Properties></cybox:Object></indicator:Observable></stix:Indicator>
	  <stix:Indicator id="example:Indicator-d81f86b9-975b-bc0b-775e-810c5ad1111" xsi:type="indicator:IndicatorType">
	    <indicator:Title>Malicious IP Addresses</indicator:Title>
	    <indicator:Type xsi:type="stixVocabs:IndicatorTypeVocab-1.0">IP Watchlist</indicator:Type>
	    <indicator:Observable>
	    <cybox:Object><cybox:Properties xsi:type="URIObj:URIObjectType" type="URL">
	    <URIObj:Value condition="Equals">178.63.174.153</URIObj:Value>
	  </cybox:Properties></cybox:Object></indicator:Observable></stix:Indicator>
	  <stix:Indicator id="example:Indicator-d81f86b9-975b-bc0b-775e-810c5ad1111" xsi:type="indicator:IndicatorType">
	    <indicator:Title>Malicious IP Addresses</indicator:Title>
	    <indicator:Type xsi:type="stixVocabs:IndicatorTypeVocab-1.0">IP Watchlist</indicator:Type>
	    <indicator:Observable>
	    <cybox:Object><cybox:Properties xsi:type="URIObj:URIObjectType" type="URL">
	    <URIObj:Value condition="Equals">188.241.140.212</URIObj:Value>
	  </cybox:Properties></cybox:Object></indicator:Observable></stix:Indicator>
	  <stix:Indicator id="example:Indicator-d81f86b9-975b-bc0b-775e-810c5ad1111" xsi:type="indicator:IndicatorType">
	    <indicator:Title>Malicious IP Addresses</indicator:Title>
	    <indicator:Type xsi:type="stixVocabs:IndicatorTypeVocab-1.0">IP Watchlist</indicator:Type>
	    <indicator:Observable>
	    <cybox:Object><cybox:Properties xsi:type="URIObj:URIObjectType" type="URL">
	    <URIObj:Value condition="Equals">14.138.73.47</URIObj:Value>
	  </cybox:Properties></cybox:Object></indicator:Observable></stix:Indicator>
	  <stix:Indicator id="example:Indicator-d81f86b9-975b-bc0b-775e-810c5ad1111" xsi:type="indicator:IndicatorType">
	    <indicator:Title>Malicious IP Addresses</indicator:Title>
	    <indicator:Type xsi:type="stixVocabs:IndicatorTypeVocab-1.0">IP Watchlist</indicator:Type>
	    <indicator:Observable>
	    <cybox:Object><cybox:Properties xsi:type="URIObj:URIObjectType" type="URL">
	    <URIObj:Value condition="Equals">131.72.138.45</URIObj:Value>
	  </cybox:Properties></cybox:Object></indicator:Observable></stix:Indicator>
	  <stix:Indicator id="example:Indicator-d81f86b9-975b-bc0b-775e-810c5ad1111" xsi:type="indicator:IndicatorType">
	    <indicator:Title>Malicious IP Addresses</indicator:Title>
	    <indicator:Type xsi:type="stixVocabs:IndicatorTypeVocab-1.0">IP Watchlist</indicator:Type>
	    <indicator:Observable>
	    <cybox:Object><cybox:Properties xsi:type="URIObj:URIObjectType" type="URL">
	    <URIObj:Value condition="Equals">62.84.51.39</URIObj:Value>
	  </cybox:Properties></cybox:Object></indicator:Observable></stix:Indicator>
	  <stix:Indicator id="example:Indicator-d81f86b9-975b-bc0b-775e-810c5ad1111" xsi:type="indicator:IndicatorType">
	    <indicator:Title>Malicious IP Addresses</indicator:Title>
	    <indicator:Type xsi:type="stixVocabs:IndicatorTypeVocab-1.0">IP Watchlist</indicator:Type>
	    <indicator:Observable>
	    <cybox:Object><cybox:Properties xsi:type="URIObj:URIObjectType" type="URL">
	    <URIObj:Value condition="Equals">62.109.23.246</URIObj:Value>
	  </cybox:Properties></cybox:Object></indicator:Observable></stix:Indicator>
	  <stix:Indicator id="example:Indicator-d81f86b9-975b-bc0b-775e-810c5ad1111" xsi:type="indicator:IndicatorType">
	    <indicator:Title>Malicious IP Addresses</indicator:Title>
	    <indicator:Type xsi:type="stixVocabs:IndicatorTypeVocab-1.0">IP Watchlist</indicator:Type>
	    <indicator:Observable>
	    <cybox:Object><cybox:Properties xsi:type="URIObj:URIObjectType" type="URL">
	    <URIObj:Value condition="Equals">5.101.113.169</URIObj:Value>
	  </cybox:Properties></cybox:Object></indicator:Observable></stix:Indicator>
	  <stix:Indicator id="example:Indicator-d81f86b9-975b-bc0b-775e-810c5ad1111" xsi:type="indicator:IndicatorType">
	    <indicator:Title>Malicious IP Addresses</indicator:Title>
	    <indicator:Type xsi:type="stixVocabs:IndicatorTypeVocab-1.0">IP Watchlist</indicator:Type>
	    <indicator:Observable>
	    <cybox:Object><cybox:Properties xsi:type="URIObj:URIObjectType" type="URL">
	    <URIObj:Value condition="Equals">213.231.8.30</URIObj:Value>
	  </cybox:Properties></cybox:Object></indicator:Observable></stix:Indicator>
	</stix:Indicators></stix:STIX_Package>`

	return rawxmldata
}
