package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/substitutes/push-cli/parser"
	"github.com/substitutes/substitutes/structs"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"path/filepath"
)

var (
	verbose   = kingpin.Flag("verbose", "Enable verbose output").Short('v').Bool()
	server    = kingpin.Flag("server", "Specify the push backend server URL").Short('s').Required().URL()
	username  = kingpin.Flag("username", "Specify the username for the backend service").Short('u').Required().String()
	password  = kingpin.Flag("password", "Specify the password for the backend service").Short('p').Required().String()
	directory = kingpin.Arg("directory", "Specify the directory to listen").Required().ExistingDir()
	proxy     = kingpin.Flag("proxy", "Enable HTTP proxy").Bool()
	proxyURL  = kingpin.Flag("proxy-url", "Proxy URL").URL()
)

func main() {
	kingpin.Parse()

	log.SetLevel(log.InfoLevel)
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}
	// api := API{}

	// api.pingAPI()

	log.Debug("Initialized application")

	if !filepath.IsAbs(*directory) {
		var err error
		*directory, err = filepath.Abs(*directory)
		if err != nil {
			log.Fatal("An error occurred while attempting to resolve the directory: ", err)
		}
	}
	log.Debugf("Using absolute directory %s! ", *directory)

	// Parse data from directory
	classesFile, err := ioutil.ReadFile(*directory + "/Druck_Kla.htm")
	if err != nil {
		log.Fatal("Failed to read classes file (Druck_Kla.htm): ", err)
	}

	// Parse classes
	classes := parser.GetClasses(classesFile[:])

	for _, class := range classes {
		classFile, err := ioutil.ReadFile(*directory + "/" + class)
		if err != nil {
			log.Fatal("Failed to read class: ", class, err)
		}
		data := parser.GetSubstitutes(classFile[:])
		// TODO: Push resulting JSON to server
		pushData(data)
	}
}

func pushData(data structs.SubstituteResponse) {}
