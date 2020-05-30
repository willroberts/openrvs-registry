package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"time"
	"unicode"
	"unicode/utf8"

	beacon "github.com/ijemafe/openrvs-beacon"
)

const (
	HealthCheckInterval  = 1 * time.Minute
	FailedCheckThreshold = 15    // Hide servers after being down 15 mins.
	PassedCheckThreshold = 2     // Show servers again after passing 2 checks.
	MaxFailedChecks      = 10080 // Prune servers from the list entirely after being down 7 days.
)

var typeMap = map[string]string{
	// Raven Shield modes
	"RGM_BombAdvMode":           "adv",  // Bomb
	"RGM_DeathmatchMode":        "adv",  // Survival
	"RGM_EscortAdvMode":         "adv",  // Pilot
	"RGM_HostageRescueAdvMode":  "adv",  // Hostage
	"RGM_HostageRescueCoopMode": "coop", // Hostage Rescue
	"RGM_HostageRescueMode":     "coop",
	"RGM_MissionMode":           "coop", // Mission
	"RGM_SquadDeathmatch":       "adv",
	"RGM_SquadTeamDeathmatch":   "adv",
	"RGM_TeamDeathmatchMode":    "adv",  // Team Survival
	"RGM_TerroristHuntCoopMode": "coop", // Terrorist Hunt
	"RGM_TerroristHuntMode":     "coop",

	// Athena Sword modes
	"RGM_CaptureTheEnemyAdvMode": "adv",
	"RGM_CountDownMode":          "coop",
	"RGM_KamikazeMode":           "adv",
	"RGM_ScatteredHuntAdvMode":   "adv",
	"RGM_TerroristHuntAdvMode":   "adv",

	// TODO: Add Iron Wrath modes
	// Free Backup, Gas Alert, Intruder, Limited Seats, Virus Upload (all adv)
}

// Temporary globals for development.
var (
	servers       map[string]server
	testservers   map[string]server
	latestversion = []byte("v1.5")
)

type server struct {
	// Fields exposed to clients.
	Name     string
	IP       string
	Port     int
	GameMode string

	// Internal fields.
	healthy      bool
	passedChecks int
	failedChecks int
}

func main() {
	servers = make(map[string]server, 0)
	testservers = make(map[string]server, 0)
	loadTestServers()

	// Start listening on UDP/8080 for beacons.
	go func() {
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
		log.Println("udp listener started")
		for {
			n, addr, err := conn.ReadFromUDP(b)
			if err != nil {
				log.Println("udp error:", err)
				continue
			}
			parseUDPMessage(addr.IP.String(), b[0:n]) // IP and message body.
		}
	}()

	// Do some UDP testing.
	go func() {
		for {
			time.Sleep(5 * time.Second)
			testUDP()
		}
	}()

	// Start listening on TCP/8080 for HTTP requests from OpenRVS clients.
	http.HandleFunc("/servers", func(w http.ResponseWriter, r *http.Request) {
		w.Write(serversToCSV(servers))
	})
	http.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Write(latestversion)
	})
	log.Println("starting http listener")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

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

//
// When we receive UDP traffic from OpenRVS Game Servers, parse the beacon,
// healthcheck the server, and update the serverlist.
//

func parseUDPMessage(ip string, msg []byte) {
	report, err := beacon.ParseServerReport(ip, msg)
	if err != nil {
		log.Println("failed to parse beacon for server", ip)
	}

	// When testing locally, key on Server Name instead of IP+Port.
	if report.IPAddress == "127.0.0.1" {
		servers[report.ServerName] = server{
			Name:     report.ServerName,
			IP:       report.IPAddress,
			Port:     report.Port,
			GameMode: typeMap[report.CurrentMode],
		}
	} else {
		servers[hostportToKey(report.IPAddress, report.Port)] = server{
			Name:     report.ServerName,
			IP:       report.IPAddress,
			Port:     report.Port,
			GameMode: typeMap[report.CurrentMode],
		}
	}

	log.Printf("there are now %d registered servers (confirm over http)", len(servers))
}

func healthcheck(s server) {
	var failed bool
	_, err := beacon.GetServerReport(s.IP, s.Port+1000)
	if err != nil {
		log.Println("healthcheck err:", err)
		failed = true
	}

	// Mark servers unhealthy after three failed healthchecks.
	if failed {
		s.passedChecks = 0
		s.failedChecks++
		if s.failedChecks >= FailedCheckThreshold {
			s.healthy = false
		}
		if s.failedChecks >= MaxFailedChecks {
			//todo: remove this server from allservers
		}
		return
	}

	// Healthcheck succeeded.
	s.passedChecks++
	s.failedChecks = 0

	// Mark unhealthy servers healthy again after three successful checks.
	if !s.healthy && s.passedChecks >= PassedCheckThreshold {
		s.healthy = true
	}
}

// Every time the app starts up, it checks the file 'checkpoint.csv' to see if
// it can pick up where it last left off. If this file does not exist, fall back
// to 'seed.csv', which contains the initial seed list for the app.
func loadSeedServers() {

}

// Temporary test data.
func loadTestServers() {
	testservers = map[string]server{
		"185.24.221.23:7777": server{Name: "SMC Suppressed Stealth",
			IP: "185.24.221.23", Port: 7777, GameMode: "coop", healthy: true},
		"208.70.251.154:7777": server{Name: "DMM Tango Hunters",
			IP: "208.70.251.154", Port: 7777, GameMode: "coop", healthy: true},
		"72.251.228.169:7777": server{Name: "OBSOLETESUPERSTARS.COM",
			IP: "72.251.228.169", Port: 7777, GameMode: "coop", healthy: true},
		"5.9.50.39:8777": server{Name: "ALLR6 | Europe TH",
			IP: "5.9.50.39", Port: 8777, GameMode: "coop", healthy: true},
		"107.172.191.114:7777": server{Name: "~24/7 Deathmatch~",
			IP: "107.172.191.114", Port: 7777, GameMode: "adv", healthy: true},
		"94.255.250.173:7777": server{Name: "|COOP|SweServer :7777",
			IP: "94.255.250.173", Port: 7777, GameMode: "coop", healthy: true},
	}

	log.Printf("loaded %d servers", len(testservers))
}

func testUDP() {
	// Pretend to be a server and replay the beacon.
	// Note: Always comes from 127.0.0.1, which breaks IP detection
	log.Printf("sending %d udp beacons", len(testservers))
	for _, s := range testservers {
		bytes, err := beacon.GetServerReport(s.IP, s.Port+1000)
		if err != nil {
			log.Println("testudp->get error:", err)
			continue
		}
		conn, err := net.DialUDP("udp4", nil,
			&net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080})
		if err != nil {
			log.Println("testudp->dial error:", err)
			continue
		}
		if _, err = conn.Write(bytes); err != nil {
			log.Println("testudp->write error:", err)
			continue
		}
	}
}

func hostportToKey(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
