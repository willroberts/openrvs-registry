package registry

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// CSVSerializer provides an interface for serializing and deserializing lists
// of OpenRVS servers as CSV bytes.
type CSVSerializer interface {
	Serialize(ServerMap) []byte
	Deserialize([]byte) (ServerMap, error)
	EnableDebug(bool)
}

// csvSerializer implements the CSVSerializer interface.
type csvSerializer struct {
	headerLine string
	debugMode  bool
}

// NewCSVSerializer initializes and returns a CSVSerializer. The debugMode
// parameter control whether or not health check status is included in
// serialized output.
func NewCSVSerializer() CSVSerializer {
	return &csvSerializer{
		headerLine: "name,ip,port,mode",
		debugMode:  false,
	}
}

func (c *csvSerializer) EnableDebug(value bool) {
	c.debugMode = value
}

// Serialize writes the given ServerMap as sorted CSV output.
func (c *csvSerializer) Serialize(m ServerMap) []byte {
	lines := []string{c.headerLine}

	var serverLines []string
	for _, server := range m {
		line := fmt.Sprintf(
			"%s,%s,%d,%s",
			server.Name,
			server.IP,
			server.Port,
			server.GameMode,
		)
		if c.debugMode {
			line += fmt.Sprintf(
				",healthy=%v,expired=%v,passed=%d,failed=%d",
				server.Health.Healthy,
				server.Health.Expired,
				server.Health.PassedChecks,
				server.Health.FailedChecks,
			)
		}
		serverLines = append(serverLines, line)
	}

	// Previous implementation sorted non-alphanumeric server names last:
	// var r rune
	// utf8.EncodeRune([]byte(line[0]), r))
	// if unicode.IsLetter(r) { ... }
	sort.Strings(serverLines)
	lines = append(lines, serverLines...)

	return []byte(strings.Join(lines, "\n"))
}

func (c *csvSerializer) Deserialize(b []byte) (ServerMap, error) {
	servers := make(ServerMap)

	lines := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	for _, line := range lines {
		// Don't attempt to deserialize the header line.
		if strings.HasPrefix(line, c.headerLine) {
			continue
		}

		// Don't attempt to deserialize malformed lines.
		fields := strings.Split(line, ",")
		if len(fields) != 4 {
			return nil, errors.New("invalid line in csv input")
		}

		// Convert port to integer.
		ip := fields[1]
		port, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, errors.New("invalid (non-numeric) port received")
		}

		// Save a new Server object in the ServerMap.
		servers[NewHostport(ip, port)] = Server{
			Name:     fields[0],
			IP:       ip,
			Port:     port,
			GameMode: strings.TrimSuffix(fields[3], "\n"),
		}
	}

	return servers, nil
}
