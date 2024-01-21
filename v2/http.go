package v2

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	beacon "github.com/willroberts/openrvs-beacon"
	v1 "github.com/willroberts/openrvs-registry"
)

func (r *registry) HandleHTTP(listenAddress v1.Hostport) error {
	http.HandleFunc("/servers", func(w http.ResponseWriter, req *http.Request) {
		w.Write(r.CSV.Serialize(filterHealthy(r.GameServerMap)))
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
		handleAdd(w, req, r.AddServer)
	})

	http.HandleFunc("/latest", func(w http.ResponseWriter, req *http.Request) {
		w.Write(v1.GetLatestReleaseVersion())
	})

	return http.ListenAndServe(string(listenAddress), nil)
}

func filterHealthy(input v1.GameServerMap) v1.GameServerMap {
	// Not yet implemented.
	return input
}

type AddHandler func(string, []byte)

func handleAdd(w http.ResponseWriter, req *http.Request, addHandler AddHandler) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	// POST body should contain a string with the pattern "ip:port".
	fields := strings.Split(string(body), ":")
	if len(fields) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ip := fields[0]
	port, err := strconv.Atoi(fields[1])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	HEALTH_CHECK_TIMEOUT := 1 * time.Minute
	// FIXME: This assumes BeaconPort is always Port+1000.
	data, err := beacon.GetServerReport(ip, port+1000, HEALTH_CHECK_TIMEOUT)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	addHandler(ip, data)
}
