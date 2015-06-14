// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package config

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

// Log Level 1 = basic system logging information, sent to STDOUT unless Enabled = true then it is logged to a file
// Log Level 2 =
// Log Level 3 = detailed debugging information like key variable changes
// Log Level 4 = function walk
// Log Level 5 = RAW packet/message decode and output

type ServerConfigType struct {
	System struct {
		Listen         string
		Prefix         string
		DbFile         string
		DbFileFullPath string
	}
	Logging struct {
		Enabled         bool
		LogLevel        int
		LogFile         string
		LogFileFullPath string
	}
	Services struct {
		Discovery  string
		Collection string
		Poll       string
	}
	Poll struct {
		Indent bool
	}
}

// --------------------------------------------------
// Load Configuration and Parse JSON File
// --------------------------------------------------

func (this *ServerConfigType) LoadConfig(filename string) {

	// Open and read configuration file
	sysConfigFileData, err := os.Open(filename)
	if err != nil {
		log.Fatalf("error opening configuration file: %v", err)
	}

	// --------------------------------------------------
	// Decode JSON configuration file
	// --------------------------------------------------
	// Use decoder instead of unmarshal so we can handle stream data
	decoder := json.NewDecoder(sysConfigFileData)
	err = decoder.Decode(this)

	if err != nil {
		log.Fatalf("error parsing configuration file %v", err)
	}

	// Lets assign the full paths to a few variables so we can use them later
	this.System.DbFileFullPath = this.System.Prefix + "/" + this.System.DbFile
	this.Logging.LogFileFullPath = this.System.Prefix + "/" + this.Logging.LogFile

	if this.Logging.LogLevel >= 5 {
		log.Printf("DEBUG-5: System Configuration Dump %+v\n", this)
	}
}

// --------------------------------------------------
// Get list of valid collections
// --------------------------------------------------

func (this *ServerConfigType) GetValidCollections() map[string]string {

	// TODO Read in from a database the collections we offer for this authenticated
	// user and put them in a map
	// TODO switch from a map to a struct so we can track more than just name and description

	// Open connection to database
	filename := this.System.DbFileFullPath
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
