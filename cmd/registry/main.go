package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	beacon "github.com/ijemafe/openrvs-beacon"
)

const (
	SeedFile       = "seed.csv"
	CheckpointFile = "checkpoint.csv"
)

var servers map[string]server

func main() {
	log.Println("loading seed servers")
	servers = make(map[string]server, 0)
	if err := loadSeedServers(); err != nil {
		log.Fatal(err)
	}
	log.Printf("there are now %d registered servers (confirm over http)", len(servers))

	// Start listening on UDP/8080 for beacons.
	go ListenUDP()

	// Do some UDP testing.
	go testUDP()

	// Start listening on TCP/8080 for HTTP requests from OpenRVS clients.
	http.HandleFunc("/servers", func(w http.ResponseWriter, r *http.Request) {
		w.Write(serversToCSV(servers))
	})
	http.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("v1.5")) // This should come from GitHub.
	})
	log.Println("starting http listener")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Every time the app starts up, it checks the file 'checkpoint.csv' to see if
// it can pick up where it last left off. If this file does not exist, fall back
// to 'seed.csv', which contains the initial seed list for the app.
func loadSeedServers() error {
	bytes, err := ioutil.ReadFile(CheckpointFile)
	if err != nil {
		log.Println("unable to read checkpoint.csv, falling back to seed.csv")
		bytes, err = ioutil.ReadFile(SeedFile)
		if err != nil {
			return err
		}
	}

	parsed, err := csvToServers(bytes)
	if err != nil {
		return err
	}
	servers = parsed
	return nil
}

// Pretend to be a server and replay the beacon. Useful for testing automatic
// addition of new servers to the list when beacons are received.
func testUDP() {
	time.Sleep(5 * time.Second)
	log.Printf("sending test udp beacon")

	// Get a real report from a test server.
	bytes, err := beacon.GetServerReport("64.225.54.237", 7777) //rs3tdm
	if err != nil {
		log.Println("testudp->get error:", err)
		return
	}
	// Connect to our own app.
	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	if err != nil {
		log.Println("testudp->dial error:", err)
		return
	}

	// Replay the report, making it appear to come from the IP 127.0.0.1
	if _, err = conn.Write(bytes); err != nil {
		log.Println("testudp->write error:", err)
		return
	}
}
