// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package responseMessages

import (
	"encoding/json"
	"github.com/freetaxii/libtaxii/discovery"
	"log"
)

// --------------------------------------------------
// Create a TAXII Discovery Response Message
// --------------------------------------------------

func CreateDiscoveryResponse(responseid string) []byte {
	pkg, m := discovery.NewResponse()
	m.AddInResponseTo(responseid)

	var s1 discovery.ServiceType
	s1.SetTypeDiscovery()
	s1.SetAvailable()
	s1.SetStandardTaxiiHttpJson()
	s1.AddAddress("http://taxiitest.freetaxii.com/services/discovery")

	var s2 discovery.ServiceType
	s2.SetTypeCollection()
	s2.SetAvailable()
	s2.SetStandardTaxiiHttpJson()
	s2.AddAddress("http://taxiitest.freetaxii.com/services/collection")

	m.AddService(s1)
	m.AddService(s2)

	data, err := json.Marshal(pkg)
	if err != nil {
		// If we can not create a status message then there is something
		// wrong with the APIs and nothing is going to work.
		log.Fatal("Unable to create Discovery Response Message")
	}
	return data
}
