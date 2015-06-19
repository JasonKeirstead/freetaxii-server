// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package taxiiserver

import (
	"encoding/json"
	"github.com/freetaxii/libtaxii/statusMessage"
	"log"
)

// --------------------------------------------------
// Create a TAXII Status Message
// --------------------------------------------------

func (this *ServerType) CreateTaxiiStatusMessage(responseid, msgType, msg string) []byte {
	tm := statusMessage.New()
	tm.AddType(msgType)
	if responseid != "" {
		tm.AddResponseId(responseid)
	}
	tm.AddMessage(msg)

	data, err := json.Marshal(tm)
	if err != nil {
		// If we can not create a status message response then there is something
		// wrong with the APIs and nothing is going to work.
		log.Fatal("Unable to create Discovery Response Message")
	}
	return data
}
