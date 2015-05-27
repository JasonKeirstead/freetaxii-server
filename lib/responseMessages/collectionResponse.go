// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package responseMessages

import (
	"encoding/json"
	"github.com/freetaxii/libtaxii/collection"
	"log"
)

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
