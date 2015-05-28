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

const (
	DEFAULT_CONFIG_FILENAME = "etc/freetaxii.conf"
	DEFAULT_LOG_FILENAME    = "logs/freetaxii.log"
)

type ConfigFileType struct {
	System struct {
		Debug   int
		LogFile string
		Listen  string
	}
	Services struct {
		Discovery  string
		Collection string
		Poll       string
	}
}

var sVersion = "0.2.1"
var DebugLevel int = 0

var sOptConfigFilename = getopt.StringLong("config", 'c', DEFAULT_CONFIG_FILENAME, "Configuration File", "string")
var sOptLogFilename = getopt.StringLong("logfile", 'f', DEFAULT_LOG_FILENAME, "Server Log File", "string")
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

	sysConfigFilename := *sOptConfigFilename
	var syscfg ConfigFileType
	err := gcfg.ReadFileInto(&syscfg, sysConfigFilename)
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
	// The default location for the logs is ./etc/freetaxii.log
	// If a log file location is passed in via the command line flags, then lets
	// use it. Otherwise, lets look in the configuration file.  If nothing is
	// there, then we will use the default.

	// TODO
	// Need to make the directory if it does not already exist
	// To do this, we need to split the filename from the directory, we will want to only
	// take the last bit in case there is multiple directories /etc/foo/bar/stuff.log

	sysLogFilename := *sOptLogFilename
	if sysLogFilename == DEFAULT_LOG_FILENAME && syscfg.System.LogFile != "" {
		sysLogFilename = syscfg.System.LogFile
	}

	logFile, err := os.OpenFile(sysLogFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Println("Starting FreeTAXII Server")

	// --------------------------------------------------
	// Setup Directory Path Handlers
	// --------------------------------------------------
	// Make sure there is a directory path defined in the configuration file
	// for each service we want to listen on.

	var taxiiServer httpHandlers.HttpHandlersType
	taxiiServer.DebugLevel = DebugLevel
	serviceCounter := 0

	if syscfg.Services.Discovery != "" {
		log.Println("TAXII Discovery services defined at:", syscfg.Services.Discovery)
		http.HandleFunc(syscfg.Services.Discovery, taxiiServer.DiscoveryServerHandler)
		serviceCounter++
	}

	if syscfg.Services.Collection != "" {
		log.Println("TAXII Collection services defined at:", syscfg.Services.Collection)
		http.HandleFunc(syscfg.Services.Collection, taxiiServer.CollectionServerHandler)
		serviceCounter++
	}

	if syscfg.Services.Poll != "" {
		log.Println("TAXII Poll services defined at:", syscfg.Services.Poll)
		http.HandleFunc(syscfg.Services.Poll, taxiiServer.PollServerHandler)
		serviceCounter++
	}

	if serviceCounter == 0 {
		log.Fatalln("No TAXII services are defined")
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
