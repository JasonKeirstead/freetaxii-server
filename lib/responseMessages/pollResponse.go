// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package responseMessages

import (
	"encoding/json"
	"github.com/freetaxii/libtaxii/poll"
	"log"
)

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
