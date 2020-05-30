// This file contains code which should live in the repo root and not this cmd.
// It's a bit of a dumping ground right now until I refactor.
package main

import (
	"fmt"
	"log"
	"net"
	"time"

	beacon "github.com/ijemafe/openrvs-beacon"
)

const (
	HealthCheckInterval  = 1 * time.Minute
	FailedCheckThreshold = 15    // Hide servers after being down 15 mins.
	PassedCheckThreshold = 2     // Show servers again after passing 2 checks.
	MaxFailedChecks      = 10080 // Prune servers from the list entirely after being down 7 days.
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

func ListenUDP() {
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
	for {
		n, addr, err := conn.ReadFromUDP(b)
		if err != nil {
			log.Println("udp error:", err)
			continue
		}
		parseUDPMessage(addr.IP.String(), b[0:n]) // IP and message body.
	}
}

// When we receive UDP traffic from OpenRVS Game Servers, parse the beacon,
// healthcheck the server, and update the serverlist.
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

func hostportToKey(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
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
