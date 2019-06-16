package main

import (
	"bytes"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"github.com/substitutes/push-cli/client"
	"github.com/substitutes/push-cli/parser"
	"github.com/substitutes/push-receiver/model"
	"github.com/substitutes/substitutes/models"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

var (
	verbose   = kingpin.Flag("verbose", "Enable verbose output").Short('v').Bool()
	server    = kingpin.Flag("server", "Specify the push backend server URL").Short('s').Required().URL()
	username  = kingpin.Flag("username", "Specify the username for the backend service").Short('u').Required().String()
	password  = kingpin.Flag("password", "Specify the password for the backend service").Short('p').Required().String()
	directory = kingpin.Arg("directory", "Specify the directory to listen").Required().ExistingDir()
	proxyURL  = kingpin.Flag("proxy", "Proxy URL").URL()
)

func main() {
	kingpin.Parse()

	log.SetLevel(log.InfoLevel)
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	http.DefaultClient = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(*proxyURL)}}

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

func pushData(data models.SubstituteResponse) {
	parsedData, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Attempting to push data: ", string(parsedData[:]))
	apiPath := (*server).String() + "/api/v1/substitute/class"
	apiReq, err := http.NewRequest("PUT", apiPath, bytes.NewReader(parsedData))
	if err != nil {
		log.Fatal("Failed to create client: ", err)
	}
	apiReq.SetBasicAuth(*username, *password)
	apiReq.Header.Set("User-Agent", client.UserAgent)
	apiReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(apiReq)
	if err != nil {
		log.Warn("Failed to push class ", data.Meta.Class, ": ", err)
		return
	}
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Failed to read backend response: ", err)
	}
	log.Debug("Response from backend: ", string(respData[:]))
	var respJSON model.APIResponse
	if json.Unmarshal(respData, &respJSON) != nil {
		log.Fatal("Failed to unmarshal backend response: ", err)
	}
	if resp.StatusCode == 201 {
		log.Info("Pushed class ", data.Meta.Class, " to backend")
	} else {
		log.Warn("Failed to push class: ", resp.Status, respJSON.Message)
	}
}
