package main

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"path/filepath"
)

var (
	verbose   = kingpin.Flag("verbose", "Enable verbose output").Short('v').Bool()
	server    = kingpin.Flag("server", "Specify the push backend server URL").Short('s').Required().URL()
	username  = kingpin.Flag("username", "Specify the username for the backend service").Short('u').Required().String()
	password  = kingpin.Flag("password", "Specify the password for the backend service").Short('p').Required().String()
	directory = kingpin.Arg("directory", "Specify the directory to listen").Required().ExistingDir()
)

func main() {
	kingpin.Parse()

	log.SetLevel(log.InfoLevel)
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}
	api := API{}

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

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Failed to start fsnotify watcher: ", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Debugf("A change occurred - syncing directory %s to %s (event: %s)", *directory, *server, event.String())
				api.pushFiles()
			case err := <-watcher.Errors:
				log.Warn("An error occurred while attempting to watch the given directory: ", err)
			}
		}
	}()

	err = watcher.Add(*directory)
	if err != nil {
		log.Fatal("Failed to add directory to watcher: ", err)
	}
	<-done
}
