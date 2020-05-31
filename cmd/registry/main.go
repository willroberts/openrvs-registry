package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	beacon "github.com/ijemafe/openrvs-beacon"
	registry "github.com/ijemafe/openrvs-registry"
)

var (
	servers             = make(map[string]registry.Server) // Stores all known servers.
	lock                = sync.RWMutex{}                   // For safely accessing the server map.
	checkpointInterval  = 5 * time.Minute                  // Save to disk this often.
	healthcheckInterval = 30 * time.Second                 // Send healthchecks this often.
)

func main() {
	log.Println("openrvs-registry process started")

	// Allow setting CSV directory explicitly. If you set this, it must include
	// a platform-dependent trailing slash. For example, on Windows:
	//     registry.exe -csvdir=C:\path\to\csv\files\\
	var dir string
	flag.StringVar(&dir, "csvdir", "", "directory containing seed.csv and checkpoint.csv")
	flag.Parse()

	// Attempt to load servers from checkpoint.csv, falling back to seed.csv.
	log.Println("loading servers from file")
	var err error
	servers, err = registry.LoadServers(dir)
	if err != nil {
		log.Fatal(err)
	}

	// Log the number of servers loaded from file.
	logServerCount()

	// Regularly checkpoint servers to disk in a new thread. This file can be
	// backed up at an OS level at regular intervals if desired.
	// The go keyword launches a goroutine, which happens concurrently and does
	// not block the current thread. For this reason, synchronization (such as
	// with lock.Lock() below) is needed.
	go func() {
		for {
			time.Sleep(checkpointInterval)
			lock.Lock()
			registry.SaveServers(dir, servers)
			lock.Unlock()
		}
	}()

	// Start listening on UDP/8080 for beacons in a new thread.
	go listenUDP()

	// Start sending healthchecks in a new thread at the configured interval.
	go func() {
		for {
			lock.Lock()
			servers = registry.SendHealthchecks(servers)
			lock.Unlock()
			time.Sleep(healthcheckInterval)
		}
	}()

	// Test automatic registration.
	// Uncomment to replay beacons to your development server.
	//go testUDP()

	// Create an HTTP handler which returns healthy servers.
	http.HandleFunc("/servers", func(w http.ResponseWriter, r *http.Request) {
		w.Write(registry.ServersToCSV(registry.FilterHealthyServers(servers), false))
	})

	// Create an HTTP handler which returns all servers.
	http.HandleFunc("/servers/all", func(w http.ResponseWriter, r *http.Request) {
		w.Write(registry.ServersToCSV(servers, false))
	})

	// Create an HTTP handler which returns all servers with detailed health status.
	http.HandleFunc("/servers/debug", func(w http.ResponseWriter, r *http.Request) {
		w.Write(registry.ServersToCSV(servers, true))
	})

	// Create an HTTP handler which returns the latest release version from Github.
	http.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Write(registry.GetLatestReleaseVersion())
	})

	// Start listening on TCP/8080 for HTTP requests from OpenRVS clients.
	log.Println("starting http listener")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// ListenUDP creates a UDP socket on 0.0.0.0:8080 and configures it to listen.
// In a blocking loop, continuously reads into a 4KB buffer, parses the source
// IP and message body, and forward the information to the registry.
func listenUDP() {
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
		log.Println("received UDP packet from", addr.IP.String())
		registerServer(addr.IP.String(), b[0:n]) // IP and message body.
	}
}

// When we receive UDP traffic from OpenRVS Game Servers, parse the beacon and
// update the serverlist.
func registerServer(ip string, msg []byte) {
	// Parses the UDP beacon from the OpenRVS server.
	report, err := beacon.ParseServerReport(ip, msg)
	if err != nil {
		log.Println("failed to parse beacon for server", ip)
	}

	// Creates and saves a Server using the beacon data.
	lock.Lock()
	servers[registry.HostportToKey(report.IPAddress, report.Port)] = registry.Server{
		Name:     report.ServerName,
		IP:       report.IPAddress,
		Port:     report.Port,
		GameMode: registry.GameTypes[report.CurrentMode],
	}
	lock.Unlock()

	// Logs the new server count.
	logServerCount()
}

// Write the server count to the console.
func logServerCount() {
	log.Printf("there are now %d registered servers (confirm over http)", len(servers))
}

// Pretend to be a server and replay the beacon. Useful for testing automatic
// addition of new servers to the list when beacons are received.
func testUDP() {
	time.Sleep(5 * time.Second)
	log.Printf("sending test udp beacon")

	// Get a real report from a test server.
	bytes, err := beacon.GetServerReport("64.225.54.237", 7777, 3*time.Second) //rs3tdm
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
