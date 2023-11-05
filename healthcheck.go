package registry

// This file contains code which should live in the repo root and not this cmd.
// It's a bit of a dumping ground right now until I refactor.

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
func SendHealthchecks(servers map[string]Server) map[string]Server {
	var (
		checked = make(map[string]Server, 0) // Output map.
		wg      sync.WaitGroup               // For synchronizing the UDP beacons.
		lock    = sync.RWMutex{}             // For safely accessing checked map.
	)

	for k, s := range servers {
		wg.Add(1) // Add an item to wait for.
		// Kick off this work in a new thread.
		go func(k string, s Server) {
			updated := UpdateHealthStatus(s) // Retrieve updated health status.
			lock.Lock()                      // Prevent other access to checked map (will block).
			checked[k] = updated             // Store the updated server info, healthy or not.
			lock.Unlock()                    // Allow other access to checked map.
			wg.Done()                        // Remove an item to wait for.
		}(k, s)
	}
	wg.Wait() // Wait until there are no items left to wait for.

	log.Println("healthy servers:", len(FilterHealthyServers(checked)), "out of", len(servers))
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
func FilterHealthyServers(servers map[string]Server) map[string]Server {
	filtered := make(map[string]Server)
	for k, s := range servers {
		if s.Health.Healthy {
			filtered[k] = s // Copy to output map.
		}
	}
	return filtered
}
