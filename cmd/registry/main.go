package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	beacon "github.com/willroberts/openrvs-beacon"
	registry "github.com/willroberts/openrvs-registry"
)

var (
	seedPath       string
	checkpointPath string

	servers             = make(map[string]registry.Server) // Stores all known servers.
	lock                = sync.RWMutex{}                   // For safely accessing the server map.
	checkpointInterval  = 5 * time.Minute                  // Save to disk this often.
	healthcheckInterval = 30 * time.Second                 // Send healthchecks this often.

	localNetworks = []string{
		"127.0.0.0/8",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}
)

func init() {
	flag.StringVar(&seedPath, "seed-file", "", "path to seed.csv")
	flag.StringVar(&checkpointPath, "checkpoint-file", "", "path to checkpoint.csv")
	flag.Parse()
}

func main() {
	log.Println("openrvs-registry process started")

	// Attempt to load servers from checkpoint.csv, falling back to seed.csv.
	log.Println("loading servers from file")
	var err error
	servers, err = registry.LoadServers(checkpointPath)
	if err != nil {
		log.Println("unable to read checkpoint.csv; falling back to seed.csv")
		servers, err = registry.LoadServers(seedPath)
		if err != nil {
			log.Fatal("unable to load servers from csv: ", err)
		}
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

	// Create an HTTP handler which accepts hints for new servers to healthcheck.
	http.HandleFunc("/servers/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		parts := strings.Split(string(body), ":")
		if len(parts) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		host := parts[0]
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		reportBytes, err := beacon.GetServerReport(host, port+1000, registry.HealthCheckTimeout)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		registerServer(host, reportBytes)
	})

	// Start listening on TCP/8080 for HTTP requests from OpenRVS clients.
	log.Println("listening on http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
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
	// Reject traffic from LAN servers.
	for _, n := range localNetworks {
		_, sub, _ := net.ParseCIDR(n)
		if sub.Contains(net.ParseIP(ip)) {
			log.Println("skipping server with local ip:", ip)
			return
		}
	}

	// Parses the UDP beacon from the OpenRVS server.
	report, err := beacon.ParseServerReport(ip, msg)
	if err != nil {
		log.Println("failed to parse beacon for server", ip)
	}

	// Validate input, checking for required fields.
	if (report.ServerName == "") || (report.Port == 0) || (report.CurrentMode == "") {
		return // Insufficient data; nothing to register.
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
