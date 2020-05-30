package main

import (
	"log"
	"net"
	"time"

	beacon "github.com/ijemafe/openrvs-beacon"
)

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
