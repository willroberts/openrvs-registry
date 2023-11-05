package main

import (
	"net"
	"testing"
	"time"

	beacon "github.com/willroberts/openrvs-beacon"
)

// Pretend to be a server and replay the beacon. Useful for testing automatic
// addition of new servers to the list when beacons are received.
func TestUDP(t *testing.T) {
	time.Sleep(5 * time.Second)
	t.Log("sending test udp beacon")

	// Get a real report from a test server.
	bytes, err := beacon.GetServerReport("184.73.85.28", 7777, 3*time.Second) //rs3tdm
	if err != nil {
		t.Log("testudp->get error:", err)
		t.FailNow()
	}
	// Connect to our own app.
	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
	if err != nil {
		t.Log("testudp->dial error:", err)
		t.FailNow()
	}

	// Replay the report, making it appear to come from the IP 127.0.0.1
	if _, err = conn.Write(bytes); err != nil {
		t.Log("testudp->write error:", err)
		t.FailNow()
	}
}
