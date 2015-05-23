// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package main

import (
	"code.google.com/p/gcfg"
	"code.google.com/p/getopt"
	"fmt"
	"github.com/freetaxii/freetaxii-server/lib/httpHandlers"
	"log"
	"net/http"
	"os"
)

type ConfigFileType struct {
	System struct {
		Debug   int
		LogFile string
	}
	Services struct {
		DiscoveryServer  string
		CollectionServer string
	}
}

var sVersion = "0.1"
var DebugLevel int = 0

//var sOptPort = getopt.StringLong("port", 'p', "8000", "Port Number (ex. 8000)", "string")
var sOptConfigFile = getopt.StringLong("config", 'c', "etc/freetaxii.conf", "Configuration File", "string")
var sOptDiscovery = getopt.StringLong("discovery-server", 0, "/services/discovery", "Service Directory for Discovery Service (ex. /services/discovery)", "string")
var sOptCollection = getopt.StringLong("collection-server", 0, "/services/collection", "Service Directory for Collection Service (ex. /services/collection)", "string")
var sOptLogFile = getopt.StringLong("logfile", 'f', "log/freetaxii.log", "Server Log File", "string")
var bOptHelp = getopt.BoolLong("help", 0, "Help")
var bOptVer = getopt.BoolLong("version", 0, "Version")

func main() {
	getopt.HelpColumn = 35
	getopt.DisplayWidth = 120
	getopt.SetParameters("")
	getopt.Parse()

	if *bOptVer {
		printOutputHeader()
		os.Exit(0)
	}

	if *bOptHelp {
		printOutputHeader()
		getopt.Usage()
		os.Exit(0)
	}

	// --------------------------------------------------
	// Load Configuration File
	// --------------------------------------------------

	sysConfigFile := *sOptConfigFile
	var syscfg ConfigFileType
	err := gcfg.ReadFileInto(&syscfg, sysConfigFile)
	if err != nil {
		log.Fatalf("error opening configuration file: %v", err)
	}

	// --------------------------------------------------
	// Setup Debug Level
	// --------------------------------------------------

	if syscfg.System.Debug >= 0 && syscfg.System.Debug <= 5 {
		DebugLevel = syscfg.System.Debug
	}

	// --------------------------------------------------
	// Setup Logging File
	// --------------------------------------------------
	// The default open for the logs in ./etc/freetaxii.log
	// If there is an option in the configuration file it will take precedence over the default
	// If a command line option is give, it will take precidence over the configuration file

	// TODO
	// Need to make the directory if it does not already exist
	// To do this, we need to split the filename from the directory, we will want to only
	// take the last bit in case there is multiple directories /etc/foo/bar/stuff.log

	var sysLogFile string
	if *sOptLogFile == "log/freetaxii.log" && syscfg.System.LogFile != "" {
		sysLogFile = syscfg.System.LogFile
	} else {
		sysLogFile = *sOptLogFile
	}

	logFile, err := os.OpenFile(sysLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Println("Starting FreeTAXII Server")

	// --------------------------------------------------
	// Setup Directory Path Handlers
	// --------------------------------------------------

	var taxiiServer httpHandlers.HttpHandlersType
	taxiiServer.DebugLevel = DebugLevel

	http.HandleFunc(*sOptDiscovery, taxiiServer.DiscoveryServerHandler)
	http.HandleFunc(*sOptCollection, taxiiServer.CollectionServerHandler)

	// --------------------------------------------------
	// Listen for Incoming Connections
	// --------------------------------------------------

	// TODO connect this up to the command line flags
	http.ListenAndServe(":8000", nil)

}

// func logfileExists(path string) (bool, error) {
//     _, err := os.Stat(path)
//     if err == nil { return true, nil }
//     if os.IsNotExist(err) { return false, nil }
//     return false, err
// }

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
