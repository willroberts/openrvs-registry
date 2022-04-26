// Send healthchecks to each server in the given CSV file.
package main

import (
	"flag"
	"log"

	registry "github.com/willroberts/openrvs-registry"
)

var csvFile string

func init() {
	flag.StringVar(&csvFile, "csv-file", "", "path to csv file containing servers")
	flag.Parse()
}

func main() {
	if csvFile == "" {
		log.Fatal("csv file must be provided")
	}

	servers, err := registry.LoadServers(csvFile)
	if err != nil {
		log.Fatal("failed to parse servers from csv:", err)
	}

	log.Println("sending healthchecks...")
	servers = registry.SendHealthchecks(servers)

	for _, s := range servers {
		h := s.Health
		if !h.Healthy {
			continue
		}
		log.Printf("Server: %s (%s:%d)", s.Name, s.IP, s.Port)
		log.Println("        ", h.Healthy, h.Expired, h.PassedChecks, h.FailedChecks)
	}
}
