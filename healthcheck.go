package registry

import (
	"log"
	"sync"
	"time"

	beacon "github.com/willroberts/openrvs-beacon"
)

const (
	// HealthCheckTimeout is the amount of time to wait before closing the UDP socket.
	HealthCheckTimeout = 5 * time.Second // Values below 3 lose data.

	// FailedCheckThreshold is used to hide servers after failing healthchecks.
	FailedCheckThreshold = 60 // 30 minutes

	// PassedCheckThreshold is used to show servers again after being marked unhealthy.
	PassedCheckThreshold = 1

	// MaxFailedChecks is used to prune servers from the list entirely.
	MaxFailedChecks = 5760 // 2 days
)

// SendHealthchecks queries each known server and updates its health status in
// memory.
// TODO: Automatically prune servers which have failed healthchecks for a week.
// This would be a bit easier if we stored the "last passed check" timestamp in
// the CSV file, since restarting the service currently restarts the number of
// consecutive failed checks (and we need over 20,000 failed checks to
// constitute a week).
func SendHealthchecks(servers ServerMap) ServerMap {
	var (
		checked = make(ServerMap, 0) // Output map.
		wg      sync.WaitGroup       // For synchronizing the UDP beacons.
		lock    = sync.RWMutex{}     // For safely accessing checked map.
	)

	for k, s := range servers {
		// Kick off this work in a new thread.
		wg.Add(1)
		go func(k Hostport, s Server) {
			updated := UpdateHealthStatus(s)
			lock.Lock()
			checked[k] = updated
			lock.Unlock()
			wg.Done()
		}(k, s)
	}
	wg.Wait()

	// TODO: Replace log spam with Prometheus counter.
	//log.Println("healthy servers:", len(FilterHealthyServers(checked)), "out of", len(servers))
	return checked
}

// UpdateHealthStatus modifies and returns a Server based on a healthcheck
// result.
func UpdateHealthStatus(s Server) Server {
	// Send a UDP beacon and determine if it failed.
	var failed bool
	reportBytes, err := beacon.GetServerReport(s.IP, s.Port+1000, HealthCheckTimeout)
	if err != nil {
		failed = true // No need to log connection refused, timeout, etc.
	}

	// Healthcheck failed.
	if failed {
		s.Health.PassedChecks = 0 // 0 checks in a row have passed
		s.Health.FailedChecks++   // Another check in a row has failed
		if s.Health.FailedChecks == FailedCheckThreshold {
			log.Println("server is now unhealthy:", s.IP, s.Port)
			s.Health.Healthy = false // Too many failed checks in a row.
		}
		if s.Health.FailedChecks >= MaxFailedChecks {
			s.Health.Expired = true // TODO: Prune expired servers.
		}
		return s
	}

	// Healthcheck succeeded.
	s.Health.PassedChecks++   // Another check in a row has passed.
	s.Health.FailedChecks = 0 // 0 checks in a row have failed.

	// Update name and game mode in case they have changed.
	report, err := beacon.ParseServerReport(s.IP, reportBytes)
	if err != nil {
		log.Println("failed to parse server report when updating name and game mode:", err)
	} else {
		s.Name = report.ServerName
		s.GameMode = report.CurrentMode
	}

	// Mark unhealthy servers healthy again after three successful checks.
	if !s.Health.Healthy && s.Health.PassedChecks >= PassedCheckThreshold {
		s.Health.Healthy = true // Server is healthy again.
		log.Println("server is now healthy:", s.IP, s.Port)
	}

	return s
}

// FilterHealthyServers iterates through a map of Servers, and returns a subset
// maps of Servers which only contains the servers marked as being healthy.
func FilterHealthyServers(servers ServerMap) ServerMap {
	filtered := make(ServerMap)
	for k, s := range servers {
		if s.Health.Healthy {
			filtered[k] = s // Copy to output map.
		}
	}
	return filtered
}
