package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	SeedFile       = "seed.csv"
	CheckpointFile = "checkpoint.csv"
)

var checkpointInterval = 5 * time.Minute

// Converts our internal data to CSV format for OpenRVS clients.
// Also handles sorting.
func serversToCSV(servers map[string]server) []byte {
	var alphaServers []string
	var nonalphaServers []string

	resp := "name,ip,port,mode\n" // CSV header line.

	for _, s := range servers {
		// Encode first letter of server name for sorting purposes.
		var r rune
		line := fmt.Sprintf("%s,%s,%d,%s", s.Name, s.IP, s.Port, s.GameMode)
		utf8.EncodeRune([]byte{line[0]}, r)

		if unicode.IsLetter(r) {
			alphaServers = append(alphaServers, line)
		} else {
			nonalphaServers = append(nonalphaServers, line)
		}
	}

	sort.Strings(alphaServers)
	sort.Strings(nonalphaServers)
	allservers := append(alphaServers, nonalphaServers...)

	for i, s := range allservers {
		resp += s
		if i != len(allservers)-1 {
			resp += "\n"
		}
	}

	return []byte(resp)
}

func csvToServers(csv []byte) (map[string]server, error) {
	servers := make(map[string]server, 0)
	lines := strings.Split(string(csv), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ",")
		port, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, err
		}
		s := server{
			Name:     fields[0],
			IP:       fields[1],
			Port:     port,
			GameMode: fields[3],
		}
		servers[hostportToKey(s.IP, s.Port)] = s
	}
	return servers, nil
}

// Every time the app starts up, it checks the file 'checkpoint.csv' to see if
// it can pick up where it last left off. If this file does not exist, fall back
// to 'seed.csv', which contains the initial seed list for the app.
func loadSeedServers() error {
	bytes, err := ioutil.ReadFile(CheckpointFile)
	if err != nil {
		log.Println("unable to read checkpoint.csv, falling back to seed.csv")
		bytes, err = ioutil.ReadFile(SeedFile)
		if err != nil {
			return err
		}
	}

	parsed, err := csvToServers(bytes)
	if err != nil {
		return err
	}
	servers = parsed
	return nil
}

func saveCheckpoint(servers map[string]server) error {
	return ioutil.WriteFile(CheckpointFile, serversToCSV(servers), 0644)
}
