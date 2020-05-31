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

// ServersToCSV converts our internal data to CSV format for OpenRVS clients.
// Also handles sorting, with special characters coming after alphabeticals.
// If debug is true, includes detailed health status in the response.
func ServersToCSV(servers map[string]Server, debug bool) []byte {
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

// CSVToServers converts CSV (generally from local file) to a map of servers.
func CSVToServers(csv []byte) (map[string]Server, error) {
	servers := make(map[string]Server, 0)
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
			Health:   HealthStatus{Healthy: true},
		}
		servers[HostportToKey(s.IP, s.Port)] = s
	}

	return servers, nil
}

// LoadServers reads a CSV file from disk.
// Every time the app starts up, it checks the file 'checkpoint.csv' to see if
// it can pick up where it last left off. If this file does not exist, fall back
// to 'seed.csv', which contains the initial seed list for the app.
func LoadServers(dir string) (map[string]Server, error) {
	// First, try to read checkpoint file.
	p := getPath(dir, checkpointFile)
	log.Println("reading checkpoint file at", p)
	bytes, err := ioutil.ReadFile(p)
	if err != nil {
		// Fall back to seed file.
		log.Println("unable to read checkpoint.csv, falling back to seed.csv")
		p = getPath(dir, seedFile)
		log.Println("reading seed file at", p)
		bytes, err = ioutil.ReadFile(p)
		if err != nil {
			// No file was found, return the error.
			return nil, err
		}
	}

	// Parse and return the CSV file.
	parsed, err := CSVToServers(bytes)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

// SaveServers writes the latest servers to disk.
func SaveServers(dir string, servers map[string]Server) error {
	// Write current servers to checkpoint file.
	p := getPath(dir, checkpointFile)
	log.Println("saving checkpoint file to", p)
	return ioutil.WriteFile(p, ServersToCSV(servers, false), 0644)
}

// getPath formats a local file path to a CSV config file. If dir is provided,
// the path will be prepended. If dir is excluded, the path will be treated as
// if it's in the current working directory.
func getPath(dir string, file string) string {
	if dir != "" {
		return fmt.Sprintf("%s%s", dir, file)
	}
	return file
}
