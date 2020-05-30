package main

import (
	"log"
	"net"
	"net/http"
	"time"

	beacon "github.com/ijemafe/openrvs-beacon"
)

var servers = make(map[string]server, 0) // Stores the current list of servers in memory.

func main() {
	log.Println("loading seed servers")
	if err := loadSeedServers(); err != nil {
		log.Fatal(err)
	}
	log.Printf("there are now %d registered servers (confirm over http)", len(servers))

	// Regularly checkpoint servers to disk at checkpoint.csv. This file can be
	// backed up at an OS level at regular intervals if desired.
	go func() {
		for {
			time.Sleep(checkpointInterval)
			saveCheckpoint(servers)
		}
	}()

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
