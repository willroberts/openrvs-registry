// Temporary entrypoint for development. Convert to proper entrypoint later by
// moving libraries into package registry.
package main

import (
	"fmt"
	"log"
	"net/http"
)

// Temporary globals for development.
var (
	goodservers   []server
	allservers    []server
	latestversion []byte = []byte("v1.5")
)

type server struct {
	Name     string
	IP       string
	Port     int
	GameMode string
}

func main() {
	loadTestServers()

	http.HandleFunc("/servers", func(w http.ResponseWriter, r *http.Request) {
		w.Write(serversToCSV(goodservers))
	})
	http.HandleFunc("/servers/all", func(w http.ResponseWriter, r *http.Request) {
		w.Write(serversToCSV(allservers))
	})
	http.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Write(latestversion)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serversToCSV(servers []server) []byte {
	resp := "name,ip,port,mode\n"
	for i, s := range servers {
		resp += fmt.Sprintf("%s,%s,%d,%s", s.Name, s.IP, s.Port, s.GameMode)
		if i != len(servers)-1 {
			resp += "\n"
		}
	}
	return []byte(resp)
}

// Temporary test data.
func loadTestServers() {
	goodservers = []server{
		server{Name: "SMC Suppressed Stealth", IP: "185.24.221.23", Port: 7777, GameMode: "coop"},
		server{Name: "DMM Tango Hunters", IP: "208.70.251.154", Port: 7777, GameMode: "coop"},
		server{Name: "OBSOLETESUPERSTARS.COM", IP: "72.251.228.169", Port: 7777, GameMode: "coop"},
	}
	allservers = append(goodservers, []server{
		server{Name: "ALLR6 | Europe TH", IP: "5.9.50.39", Port: 8777, GameMode: "coop"},
		server{Name: "~24/7 Deathmatch~", IP: "107.172.191.114", Port: 7777, GameMode: "adv"},
		server{Name: "|COOP|SweServer :7777", IP: "94.255.250.173", Port: 7777, GameMode: "coop"},
	}...)
}
