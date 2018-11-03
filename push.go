package main

import "gopkg.in/alecthomas/kingpin.v2"

var (
	verbose  = kingpin.Flag("verbose", "Enable verbose output").Short('v').Bool()
	server   = kingpin.Flag("server", "Specify the push backend server").Short('s').String()
	username = kingpin.Flag("username", "Specify the username for the backend service").Short('u').String()
	password = kingpin.Flag("password", "Specify the password for the backend service").Short('p').String()
)

func main() {
	kingpin.Parse()

	// Verify connection against server
	
}
