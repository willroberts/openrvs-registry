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
	// SeedFile contains the initial server list to use.
	SeedFile = "seed.csv"
	// CheckpointFile contains the latest server list, changing over time.
	CheckpointFile = "checkpoint.csv"

	csvHeaderLine = "name,ip,port,mode"
)

// ServersToCSV converts our internal data to CSV format for OpenRVS clients.
// Also handles sorting, with special characters coming after alphabeticals.
func ServersToCSV(servers map[string]Server) []byte {
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

// CSVToServers converts CSV (generally from local file) to a map of servers.
func CSVToServers(csv []byte) (map[string]Server, error) {
	servers := make(map[string]Server, 0)
	trimmed := strings.TrimSuffix(string(csv), "\n")
	lines := strings.Split(trimmed, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, csvHeaderLine) {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) != 4 {
			log.Println("warning: invalid line skipped:", line)
		}
		port, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, err
		}
		s := Server{
			Name:     fields[0],
			IP:       fields[1],
			Port:     port,
			GameMode: fields[3],
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
	p := getPath(dir, CheckpointFile)
	log.Println("reading checkpoint file at", p)
	bytes, err := ioutil.ReadFile(p)
	if err != nil {
		log.Println("unable to read checkpoint.csv, falling back to seed.csv")
		p = getPath(dir, SeedFile)
		log.Println("reading seed file at", p)
		bytes, err = ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}
	}

	parsed, err := CSVToServers(bytes)
	if err != nil {
		return nil, err
	}

	return parsed, nil
}

// SaveServers writes the latest servers to disk.
func SaveServers(dir string, servers map[string]Server) error {
	p := getPath(dir, CheckpointFile)
	log.Println("saving checkpoint file to", p)
	return ioutil.WriteFile(p, ServersToCSV(servers), 0644)
}

func getPath(dir string, file string) string {
	if dir != "" {
		file = fmt.Sprintf("%s%s", dir, file)
	}
	return file
}
