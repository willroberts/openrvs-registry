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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("server added successfully"))
	})

	http.HandleFunc("/add-server", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(getFormHtml()))
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

func getFormHtml() string {
	return `
<html>
 <head>
  <title>OpenRVS.org | Add Server</title>
  <script>
   function handleButtonClick() {
    var ip_textbox = document.getElementById('ip');
    var port_textbox = document.getElementById('port');
    fetch("/servers/add", {
     method: "POST",
     body: ` + "`${ip_textbox.value}:${port_textbox.value}`" + `,
    })
    ip_textbox.value = "";
    port_textbox.value = "";
    document.getElementById('submitted_text').hidden = false;
   }
  </script>
 </head>
 <body>
  <form action="/add-server" method="POST">
   <p>
    <label for="ip">IP Address:</label>
    <input type="text" id="ip" name="ip_address" />
   </p>
   <p>
    <label for="port">Port:</label>
    <input type="text" id="port" name="port" />
   </p>
   <p class="button">
    <button type="button" onclick="handleButtonClick()">Submit</button>
    <label id="submitted_text" for="submitted" hidden=true>Submitted! Your server should appear at https://openrvs.org/servers within five minutes.</label>
   </p>
  </form>
 </body>
</html>
`
}
