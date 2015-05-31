// Copyright 2015 Bret Jordan, All rights reserved.
//
// Use of this source code is governed by an Apache 2.0 license
// that can be found in the LICENSE file in the root of the source
// tree.

package main

import (
	"bufio"
	"code.google.com/p/getopt"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"strings"
)

const (
	DEFAULT_CONFIG_FILENAME = "../etc/freetaxii.conf"
)

type ConfigFileType struct {
	System struct {
		DebugLevel      int
		Prefix          string
		LogFile         string
		DbFile          string
		LogFileFullPath string
		DbFileFullPath  string
	}
}

var sVersion = "0.2.1"
var DebugLevel int = 0

var sOptConfigFilename = getopt.StringLong("config", 'c', DEFAULT_CONFIG_FILENAME, "Configuration File", "string")
var bOptListCollection = getopt.BoolLong("list-collections", 0, "List Collections")
var bOptAddCollection = getopt.BoolLong("add-collection", 0, "Add Collections")
var bOptDelCollection = getopt.BoolLong("del-collection", 0, "Delete Collections")
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
	// Load Configuration File
	// --------------------------------------------------

	sysConfigFilename := *sOptConfigFilename
	sysConfigFileData, err := os.Open(sysConfigFilename)
	if err != nil {
		log.Fatalf("error opening configuration file: %v", err)
	}

	// --------------------------------------------------
	// Decode JSON configuration file
	// --------------------------------------------------
	// Use decoder instead of unmarshal so we can handle stream data
	decoder := json.NewDecoder(sysConfigFileData)
	var syscfg ConfigFileType
	err = decoder.Decode(&syscfg)

	if err != nil {
		log.Fatalf("error parsing configuration file %v", err)
	}

	// Lets assign the full paths to a few variables so we can use them later
	syscfg.System.DbFileFullPath = syscfg.System.Prefix + "/" + syscfg.System.DbFile
	syscfg.System.LogFileFullPath = syscfg.System.Prefix + "/" + syscfg.System.LogFile

	// --------------------------------------------------
	// Setup Debug Level
	// --------------------------------------------------

	if syscfg.System.DebugLevel >= 0 && syscfg.System.DebugLevel <= 5 {
		DebugLevel = syscfg.System.DebugLevel
	}

	// --------------------------------------------------
	// Setup Logging File
	// --------------------------------------------------
	// The default location for the logs is ./log/freetaxii.log
	// If a log file location is passed in via the command line flags, then lets
	// use it. Otherwise, lets look in the configuration file.  If nothing is
	// there, then we will use the default.

	// TODO
	// Need to make the directory if it does not already exist
	// To do this, we need to split the filename from the directory, we will want to only
	// take the last bit in case there is multiple directories /etc/foo/bar/stuff.log

	sysLogFilename := syscfg.System.LogFileFullPath

	if DebugLevel >= 3 {
		log.Println("Mgmt - Using the following log file", sysLogFilename)
	}

	logFile, err := os.OpenFile(sysLogFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Println("Starting FreeTAXII Management")

	// --------------------------------------------------
	// Open connection to database
	// --------------------------------------------------
	filename := syscfg.System.DbFileFullPath
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatalf("Unable to open file %s due to error %v", filename, err)
	}
	defer db.Close()

	if DebugLevel >= 3 {
		log.Println("Mgmt - Using the following database file", filename)
	}

	// --------------------------------------------------
	// Check for what to do
	// --------------------------------------------------
	if *bOptListCollection {
		listCollections(db)
	}
	if *bOptAddCollection {
		addCollection(db)
	}
	if *bOptDelCollection {
		delCollection(db)
	}

}

// --------------------------------------------------
// List currently defined collections
// --------------------------------------------------

func listCollections(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM Collections")
	if err != nil {
		log.Printf("Mgmt - error running query, %v", err)
	}
	defer rows.Close()

	fmt.Println("\nCurrent Collections")
	fmt.Println("===================")
	for rows.Next() {
		var collection string
		var description string
		err = rows.Scan(&collection, &description)
		if err != nil {
			log.Printf("Mgmt - error reading from database, %v", err)
		}
		fmt.Printf("\t%-10s \t %s\n", collection, description)
	}
}

// --------------------------------------------------
// Add collection
// --------------------------------------------------

func addCollection(db *sql.DB) {
	fmt.Print("Collection Name: ")
	collectionName, _ := getInput()

	fmt.Print("Collection Description: ")
	collectionDescription, _ := getInput()

	_, err := db.Exec("INSERT INTO Collections (collection, description) values (?, ?)", collectionName, collectionDescription)
	if err != nil {
		log.Printf("Mgmt - Unable to insert record due to error %v", err)
	}

	if DebugLevel >= 1 {
		log.Printf("Mgmt - Inserted %s in to table Collections", collectionName)
	}
}

// --------------------------------------------------
// Delete collection
// --------------------------------------------------
func delCollection(db *sql.DB) {
	fmt.Print("Collection Name: ")
	collectionName, _ := getInput()

	_, err := db.Exec("DELETE FROM Collections where (collection=?)", collectionName)
	if err != nil {
		log.Printf("Mgmt - Unable to delete record due to error %v", err)
	}

	// TODO this does not work right if the value is not in the database. It says it was deleted
	// when it was not, need to catch that error
	if DebugLevel >= 1 {
		log.Printf("Mgmt - Deleted %s from table Collections", collectionName)
	}
}

// --------------------------------------------------
// Get Input
// --------------------------------------------------

func getInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	return input, err
}

// --------------------------------------------------
// Print Help
// --------------------------------------------------

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
