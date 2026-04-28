package main

import (
	"flag"
	"log"
	"net"
	"time"
	"fmt"

	"github.com/willroberts/openrvs-registry/registry"
)

var (
	seedPath       string
	checkpointPath string
	listenAddress  string
	listenPort     int
	beaconPort     int
)

func init() {
	flag.StringVar(&seedPath, "seed-file", "seed.csv", "path to seed csv file, defaults to seed.csv")
	flag.StringVar(&checkpointPath, "checkpoint-file", "checkpoint.csv", "path to checkpoint csv file, defaults to checkpoint.csv")
	flag.IntVar(&listenPort, "listen-port", 8080, "port server will listen on, defaults to 8080")
	flag.StringVar(&listenAddress, "listen-address", "0.0.0.0", "ip address server will listen on, defaults to 0.0.0.0")
	flag.IntVar(&beaconPort, "beacon-port", 8080, "port server will listen for registrations on, defaults to 8080")
	flag.Parse()
}

func main() {
	log.Println("openrvs-registry process started")

	config := registry.Config{
		SeedPath:                      seedPath,
		CheckpointPath:                checkpointPath,
		CheckpointInterval:            5 * time.Minute,
		HealthcheckInterval:           30 * time.Second,
		HealthcheckTimeout:            5 * time.Second,
		HealthcheckHealthyThreshold:   1,
		HealthcheckUnhealthyThreshold: 60,   // 30 minutes.
		HealthcheckHiddenThreshold:    5760, // 2 days.
		ListenAddr:                    fmt.Sprintf("%s:%d", listenAddress, listenPort),
	}

	reg := registry.NewRegistry(config)

	// Attempt to load servers from checkpoint.csv, falling back to seed.csv.
	log.Printf("loading checkpoint servers from \"%s\"", config.CheckpointPath)
	if err := reg.LoadServers(config.CheckpointPath); err != nil {
		log.Println("failed to read checkpoint file", err)
		log.Printf("falling back to seed file from \"%s\"", config.SeedPath)
		if err := reg.LoadServers(config.SeedPath); err != nil {
			log.Println("failed to read seed file", err)
			log.Println("falling back to empty server list")
		}
	}

	// Log the number of servers loaded from file.
	log.Printf("there are now %d registered servers", reg.ServerCount())

	// Regularly checkpoint servers to disk in a new thread. This file can be
	// backed up at an OS level at regular intervals if desired.
	// The go keyword launches a goroutine, which happens concurrently and does
	// not block the current thread. For this reason, synchronization (such as
	// with lock.Lock() below) is needed.
	go func() {
		for {
			time.Sleep(config.CheckpointInterval)
			log.Printf("saving checkpoint file to \"%s\"", config.CheckpointPath)
			if err := reg.SaveServers(config.CheckpointPath); err != nil {
				log.Println("failed to write checkpoint file:", err)
			}
		}
	}()

	// Start listening on UDP/beaconPort for beacons in a new thread.
	udpHandler := func(addr *net.UDPAddr, data []byte, err error) {
		log.Println("received UDP from", addr.IP.String())
		if err != nil {
			log.Println("udp error:", err)
			return
		}
		if err := reg.AddServer(addr.IP.String(), data); err != nil {
			log.Println("registration error:", err)
			return
		}
		log.Printf("there are now %d registered servers", reg.ServerCount())
	}
	stopCh := make(chan struct{})
	log.Printf("listening on udp://0.0.0.0:%d", beaconPort)
	go reg.HandleUDP(beaconPort, udpHandler, stopCh)

	// Start sending healthchecks in a new thread at the configured interval.
	go func() {
		log.Printf("sending healthchecks every %d seconds", config.HealthcheckInterval/time.Second)
		for {
			reg.SendHealthchecks(
				// onHealthy
				func(s registry.GameServer) {
					log.Println("server is now healthy:", s.IP, s.Port)
				},
				// onUnhealthy
				func(s registry.GameServer) {
					log.Println("server is now unhealthy:", s.IP, s.Port)
				},
			)
			time.Sleep(config.HealthcheckInterval)
		}
	}()

	// Start listening for HTTP requests from OpenRVS clients.
	log.Printf("listening on http://%s", config.ListenAddr)
	log.Fatal(reg.HandleHTTP(config.ListenAddr))
}
