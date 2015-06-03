// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package main

import (
	"code.google.com/p/getopt"
	"fmt"
	"github.com/freetaxii/freetaxii-server/lib/config"
	// "github.com/freetaxii/freetaxii-server/lib/services/collection"
	"github.com/freetaxii/freetaxii-server/lib/services/discovery"
	// "github.com/freetaxii/freetaxii-server/lib/services/poll"
	"log"
	"net/http"
	"os"
)

const (
	DEFAULT_CONFIG_FILENAME = "etc/freetaxii.conf"
)

var sVersion = "0.2.1"

var sOptConfigFilename = getopt.StringLong("config", 'c', DEFAULT_CONFIG_FILENAME, "Configuration File", "string")
var bOptHelp = getopt.BoolLong("help", 0, "Help")
var bOptVer = getopt.BoolLong("version", 0, "Version")

func main() {
	getopt.HelpColumn = 35
	getopt.DisplayWidth = 120
	getopt.SetParameters("")
	getopt.Parse()

	if *bOptVer {
		printVersion()
	}

	if *bOptHelp {
		printHelp()
	}

	// --------------------------------------------------
	// Load System Configuration
	// --------------------------------------------------

	var syscfg config.ServerConfigType
	syscfg.LoadConfig(*sOptConfigFilename)

	// --------------------------------------------------
	// Setup Logging File
	// --------------------------------------------------
	// TODO
	// Need to make the directory if it does not already exist
	// To do this, we need to split the filename from the directory, we will want to only
	// take the last bit in case there is multiple directories /etc/foo/bar/stuff.log

	// Only enable logging to a file if it is turned on in the configuration file
	if syscfg.Logging.Enabled == true {
		logFile, err := os.OpenFile(syscfg.Logging.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer logFile.Close()

		log.SetOutput(logFile)
	}

	// --------------------------------------------------
	// Setup Directory Path Handlers
	// --------------------------------------------------
	// Make sure there is a directory path defined in the configuration file
	// for each service we want to listen on.
	log.Println("Starting FreeTAXII Server")
	serviceCounter := 0

	var taxiiDiscoveryServer discovery.DiscoveryType
	taxiiDiscoveryServer.SysConfig = &syscfg

	taxiiDiscoveryServer.ReloadServices = true
	if syscfg.Logging.LogLevel >= 3 {
		log.Println("DEBUG: Setting reload services to true")
	}

	if syscfg.Services.Discovery != "" {
		log.Println("Starting TAXII Discovery services at:", syscfg.Services.Discovery)
		http.HandleFunc(syscfg.Services.Discovery, taxiiDiscoveryServer.DiscoveryServerHandler)
		serviceCounter++
	}

	// var taxiiCollectionServer collection.CollectionType
	// taxiiCollectionServer.LogLevel = LogLevel
	// taxiiCollectionServer.DbFileFullPath = syscfg.System.DbFileFullPath

	// if syscfg.Services.Collection != "" {
	// 	log.Println("Starting TAXII Collection services at:", syscfg.Services.Collection)
	// 	http.HandleFunc(syscfg.Services.Collection, taxiiCollectionServer.CollectionServerHandler)
	// 	serviceCounter++
	// }

	// var taxiiPollServer poll.PollType
	// taxiiPollServer.LogLevel = LogLevel
	// taxiiPollServer.DbFileFullPath = syscfg.System.DbFileFullPath

	// if syscfg.Services.Poll != "" {
	// 	log.Println("Starting TAXII Poll services at:", syscfg.Services.Poll)
	// 	http.HandleFunc(syscfg.Services.Poll, taxiiPollServer.PollServerHandler)
	// 	serviceCounter++
	// }

	if serviceCounter == 0 {
		log.Fatalln("No TAXII services defined")
	}

	// --------------------------------------------------
	// Listen for Incoming Connections
	// --------------------------------------------------

	// TODO - Need to verify the list address is a valid IPv4 address and port
	// combination.
	if syscfg.System.Listen != "" {
		http.ListenAndServe(syscfg.System.Listen, nil)
	} else {
		log.Fatalln("The listen directive is missing from the configuration file")
	}

}

// func logfileExists(path string) (bool, error) {
//     _, err := os.Stat(path)
//     if err == nil { return true, nil }
//     if os.IsNotExist(err) { return false, nil }
//     return false, err
// }

func printHelp() {
	printOutputHeader()
	getopt.Usage()
	os.Exit(0)
}

func printVersion() {
	printOutputHeader()
	os.Exit(0)
}

// --------------------------------------------------
// Print a header for all output
// --------------------------------------------------

func printOutputHeader() {
	fmt.Println("")
	fmt.Println("FreeTAXII Server")
	fmt.Println("Copyright, Bret Jordan")
	fmt.Println("Version:", sVersion)
	fmt.Println("")
}
