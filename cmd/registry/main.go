package main

import (
	"flag"
	"log"
	"net"
	"time"

	"github.com/willroberts/openrvs-registry/registry"
)

var (
	seedPath       string
	checkpointPath string
)

func init() {
	flag.StringVar(&seedPath, "seed-file", "", "path to seed.csv")
	flag.StringVar(&checkpointPath, "checkpoint-file", "", "path to checkpoint.csv")
	flag.Parse()
}

func main() {
	log.Println("openrvs-registry process started")

	config := registry.RegistryConfig{
		SeedPath:                      seedPath,
		CheckpointPath:                checkpointPath,
		CheckpointInterval:            5 * time.Minute,
		HealthcheckInterval:           30 * time.Second,
		HealthcheckTimeout:            5 * time.Second,
		HealthcheckHealthyThreshold:   1,
		HealthcheckUnhealthyThreshold: 60,   // 30 minutes.
		HealthcheckHiddenThreshold:    5760, // 2 days.
		ListenAddr:                    "127.0.0.1:8080",
	}

	reg := registry.NewRegistry(config)

	// Attempt to load servers from checkpoint.csv, falling back to seed.csv.
	log.Println("loading servers from file")
	if err := reg.LoadServers(config.CheckpointPath); err != nil {
		log.Println("unable to read checkpoint.csv, falling back to seed.csv")
		if err := reg.LoadServers(config.SeedPath); err != nil {
			log.Println("unable to read seed.csv, falling back to empty server list")
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
			log.Println("saving checkpoint file to ", config.CheckpointPath)
			if err := reg.SaveServers(config.CheckpointPath); err != nil {
				log.Println("failed to write checkpoint file:", err)
			}
		}
	}()

	// Start listening on UDP/8080 for beacons in a new thread.
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
	log.Println("listening on udp://0.0.0.0:8080")
	go reg.HandleUDP(8080, udpHandler, stopCh)

	// Start sending healthchecks in a new thread at the configured interval.
	go func() {
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
