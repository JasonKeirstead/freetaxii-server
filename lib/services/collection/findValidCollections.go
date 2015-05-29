// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package collection

// import (
// 	"database/sql"
// 	_ "github.com/mattn/go-sqlite3"
// )

// --------------------------------------------------
// Get list of valid collections
// --------------------------------------------------

// TODO Read in from a database the collections we offer for this authenticated
// user and put them in a map

// The key is the collection name and the value is the description
func (this *CollectionType) GetValidCollections() map[string]string {
	c := make(map[string]string)
	c["ip-watch-list"] = "List of interesting IP addresses"
	c["url-watch-list"] = "List of interesting URL addresses"
	return c
}
