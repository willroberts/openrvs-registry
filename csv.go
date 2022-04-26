package registry

import (
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	seedFile       = "seed.csv"
	checkpointFile = "checkpoint.csv"
	csvHeaderLine  = "name,ip,port,mode"
)

// serversToCSV converts our internal data to CSV format for OpenRVS clients.
// Also handles sorting, with special characters coming after alphabeticals.
// If debug is true, includes detailed health status in the response.
func serversToCSV(servers map[string]Server, debug bool) []byte {
	// Use two lists to maintain alphabetical sorting.
	var alphaServers []string
	var nonalphaServers []string

	resp := "name,ip,port,mode\n" // CSV header line.

	for _, s := range servers {
		var r rune
		var line string

		if debug {
			line = fmt.Sprintf("%s,%s,%d,%s,healthy=%v,expired=%v,passed=%d,failed=%d",
				s.Name, s.IP, s.Port, s.GameMode, s.Health.Healthy, s.Health.Expired,
				s.Health.PassedChecks, s.Health.FailedChecks)
		} else {
			line = fmt.Sprintf("%s,%s,%d,%s", s.Name, s.IP, s.Port, s.GameMode)
		}

		// Encode first letter of server name for sorting purposes.
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

	// Serialize the server lines as a single string.
	for i, s := range allservers {
		resp += s
		if i != len(allservers)-1 {
			resp += "\n"
		}
	}

	return []byte(resp)
}

// csvToServers converts CSV (generally from local file) to a map of servers.
func csvToServers(csv []byte) (map[string]Server, error) {
	servers := make(map[string]Server)
	trimmed := strings.TrimSuffix(string(csv), "\n")

	for _, line := range strings.Split(trimmed, "\n") {
		// Skip the header line.
		if strings.HasPrefix(line, csvHeaderLine) {
			continue
		}

		// Ensure the format is correct.
		fields := strings.Split(line, ",")
		if len(fields) != 4 {
			log.Println("warning: invalid line skipped:", line)
			continue
		}

		// Convert port to integer.
		port, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, err
		}

		// Create and save the Server object.
		s := Server{
			Name:     fields[0],
			IP:       fields[1],
			Port:     port,
			GameMode: strings.TrimSuffix(fields[3], "\n"),
		}
		servers[HostportToKey(s.IP, s.Port)] = s
	}

	return servers, nil
}

// LoadServers reads a CSV file from disk.
// Every time the app starts up, it checks the file 'checkpoint.csv' to see if
// it can pick up where it last left off. If this file does not exist, fall back
// to 'seed.csv', which contains the initial seed list for the app.
func LoadServers(csvPath string) (map[string]Server, error) {
	// First, try to read checkpoint file.
	log.Println("reading checkpoint file at", csvPath)
	bytes, err := ioutil.ReadFile(csvPath)
	if err != nil {
		return nil, err
	}

	// Parse and return the CSV file.
	parsed, err := csvToServers(bytes)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

// SaveServers writes the latest servers to disk.
func SaveServers(csvPath string, servers map[string]Server) error {
	// Write current servers to checkpoint file.
	log.Println("saving checkpoint file to", csvPath)
	return ioutil.WriteFile(csvPath, serversToCSV(servers, false), 0644)
}
