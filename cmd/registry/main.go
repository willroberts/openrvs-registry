package main

import (
	"log"
	"net"
	"net/http"
	"time"

	beacon "github.com/ijemafe/openrvs-beacon"
	registry "github.com/ijemafe/openrvs-registry"
)

var (
	servers            = make(map[string]registry.Server, 0) // Stores the current list of servers in memory.
	checkpointInterval = 5 * time.Minute
)

func main() {
	log.Println("loading seed servers")
	var err error
	servers, err = registry.LoadServers()
	if err != nil {
		log.Fatal(err)
	}
	logServerCount()

	// Regularly checkpoint servers to disk at checkpoint.csv. This file can be
	// backed up at an OS level at regular intervals if desired.
	go func() {
		for {
			time.Sleep(checkpointInterval)
			log.Println("saving servers to checkpoint.csv")
			registry.SaveCheckpoint(servers)
		}
	}()

	// Start listening on UDP/8080 for beacons.
	go ListenUDP()

	// Test automatic registration.
	//go testUDP()

	// Start listening on TCP/8080 for HTTP requests from OpenRVS clients.
	http.HandleFunc("/servers", func(w http.ResponseWriter, r *http.Request) {
		w.Write(registry.ServersToCSV(servers))
	})
	http.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("v1.5")) // TODO: This should come from GitHub.
	})
	log.Println("starting http listener")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ListenUDP() {
	log.Println("starting udp listener")
	addr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	b := make([]byte, 4096)
	for {
		n, addr, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Println("udp error:", err)
			continue
		}
		ProcessUDP(addr.IP.String(), b[0:n]) // IP and message body.
	}
}

// When we receive UDP traffic from OpenRVS Game Servers, parse the beacon,
// healthcheck the server, and update the serverlist.
func ProcessUDP(ip string, msg []byte) {
	report, err := beacon.ParseServerReport(ip, msg)
	if err != nil {
		log.Println("failed to parse beacon for server", ip)
	}

	servers[registry.HostportToKey(report.IPAddress, report.Port)] = registry.Server{
		Name:     report.ServerName,
		IP:       report.IPAddress,
		Port:     report.Port,
		GameMode: registry.GameTypes[report.CurrentMode],
	}

	logServerCount()
}

func logServerCount() {
	log.Printf("there are now %d registered servers (confirm over http)", len(servers))
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
