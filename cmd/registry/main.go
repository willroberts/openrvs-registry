package main

import (
	"flag"
	"log"
	"net"
	"os"
	"sync"
	"time"

	beacon "github.com/willroberts/openrvs-beacon"
	registry "github.com/willroberts/openrvs-registry"
	v2 "github.com/willroberts/openrvs-registry/v2"
)

var (
	seedPath          string
	checkpointPath    string
	gameServerMap     = make(registry.GameServerMap) // Stores all known servers.
	gameServerMapLock = sync.RWMutex{}               // For safely accessing the server map.
	csv               = registry.NewCSVSerializer()
)

func init() {
	flag.StringVar(&seedPath, "seed-file", "", "path to seed.csv")
	flag.StringVar(&checkpointPath, "checkpoint-file", "", "path to checkpoint.csv")
	flag.Parse()
}

func main() {
	log.Println("openrvs-registry process started")

	config := v2.RegistryConfig{
		SeedPath:            seedPath,
		CheckpointPath:      checkpointPath,
		CheckpointInterval:  5 * time.Minute,
		HealthcheckInterval: 30 * time.Second,
		HealthcheckTimeout:  5 * time.Second,
		ListenAddr:          registry.Hostport("127.0.0.1", 8080),
	}

	reg := v2.NewRegistry(config)

	// Attempt to load servers from checkpoint.csv, falling back to seed.csv.
	log.Println("loading servers from file")
	var err error
	gameServerMap, err = LoadServers(config.CheckpointPath)
	if err != nil {
		log.Println("unable to read checkpoint.csv; falling back to seed.csv")
		gameServerMap, err = LoadServers(config.SeedPath)
		if err != nil {
			log.Println("Warning: Unable to load servers from csv: ", err)
			gameServerMap = make(registry.GameServerMap)
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
			time.Sleep(config.CheckpointInterval)
			gameServerMapLock.Lock()
			log.Println("Saving checkpoint file to ", config.CheckpointPath)
			if err := os.WriteFile(config.CheckpointPath, csv.Serialize(gameServerMap), 0644); err != nil {
				log.Println("Failed to write checkpoint file:", err)
			}
			gameServerMapLock.Unlock()
		}
	}()

	// Start listening on UDP/8080 for beacons in a new thread.
	log.Println("Starting UDP listener on port 8080")
	serveUDP()

	// Start sending healthchecks in a new thread at the configured interval.
	go func() {
		for {
			gameServerMapLock.Lock()
			gameServerMap = registry.SendHealthchecks(gameServerMap)
			gameServerMapLock.Unlock()
			time.Sleep(config.HealthcheckInterval)
		}
	}()

	// Start listening for HTTP requests from OpenRVS clients.
	log.Printf("Listening on http://%s", reg.Config.ListenAddr)
	log.Fatal(reg.HandleHTTP(reg.Config.ListenAddr))
}

// When we receive UDP traffic from OpenRVS Game Servers, parse the beacon and
// update the serverlist.
func registerServer(ip string, msg []byte) {
	// Reject traffic from LAN servers.
	if net.ParseIP(ip).IsPrivate() {
		log.Println("skipping server with private ip:", ip)
		return
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
	gameServerMapLock.Lock()
	gameServerMap[registry.NewHostport(report.IPAddress, report.Port)] = registry.GameServer{
		Name:     report.ServerName,
		IP:       report.IPAddress,
		Port:     report.Port,
		GameMode: registry.GameModes[report.CurrentMode],
	}
	gameServerMapLock.Unlock()

	// Logs the new server count.
	logServerCount()
}

// Write the server count to the console.
func logServerCount() {
	log.Printf("there are now %d registered servers", len(gameServerMap))
}

// LoadServers reads a CSV file from disk.
// Every time the app starts up, it checks the file 'checkpoint.csv' to see if
// it can pick up where it last left off. If this file does not exist, fall back
// to 'seed.csv', which contains the initial seed list for the app.
func LoadServers(csvPath string) (registry.GameServerMap, error) {
	// First, try to read checkpoint file.
	log.Println("reading checkpoint file at", csvPath)
	bytes, err := os.ReadFile(csvPath)
	if err != nil {
		return nil, err
	}

	// Parse and return the CSV file.
	parsed, err := csv.Deserialize(bytes)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

func serveUDP() {
	udpHandler := func(addr *net.UDPAddr, data []byte, err error) {
		log.Println("Received UDP from", addr.IP.String())
		if err != nil {
			log.Println("UDP error:", err)
			return
		}
		registerServer(addr.IP.String(), data)
	}
	stopCh := make(chan struct{})
	go v2.HandleUDP(8080, udpHandler, stopCh)
}
