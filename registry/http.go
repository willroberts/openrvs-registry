package registry

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	beacon "github.com/willroberts/openrvs-beacon"
	"github.com/willroberts/openrvs-registry/github"
)

func (r *registry) HandleHTTP(listenAddress string) error {
	http.HandleFunc("/latest", func(w http.ResponseWriter, req *http.Request) {
		w.Write(github.GetLatestReleaseVersion())
	})

	http.HandleFunc("/servers", func(w http.ResponseWriter, req *http.Request) {
		w.Write(r.CSV.Serialize(filterHealthyServers(r.GameServerMap)))
	})

	http.HandleFunc("/servers/all", func(w http.ResponseWriter, req *http.Request) {
		w.Write(r.CSV.Serialize(r.GameServerMap))
	})

	http.HandleFunc("/servers/debug", func(w http.ResponseWriter, req *http.Request) {
		r.CSV.EnableDebug(true)
		w.Write(r.CSV.Serialize(r.GameServerMap))
		r.CSV.EnableDebug(false)
	})

	http.HandleFunc("/servers/add", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("request method must be POST"))
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to read request body"))
			return
		}
		defer req.Body.Close()

		// POST body should contain a string with the pattern "ip:port".
		fields := strings.Split(string(body), ":")
		if len(fields) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("request body must contain 'ip:port'"))
			return
		}

		ip := fields[0]
		port, err := strconv.Atoi(fields[1])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("port must be a number"))
			return
		}

		beaconPort := port + 1000
		data, err := beacon.GetServerReport(ip, beaconPort, r.Config.HealthcheckTimeout)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to reach new server; ensure ServerBeaconPort is Port+1000 in RavenShield.ini"))
			return
		}

		r.AddServer(ip, data)
	})

	return http.ListenAndServe(string(listenAddress), nil)
}

func filterHealthyServers(servers GameServerMap) GameServerMap {
	filtered := make(GameServerMap)
	for k, s := range servers {
		if s.Health.Healthy {
			filtered[k] = s
		}
	}
	return filtered
}
