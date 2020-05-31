// This file contains code which should live in the repo root and not this cmd.
// It's a bit of a dumping ground right now until I refactor.
package registry

import (
	"log"
	"time"

	beacon "github.com/ijemafe/openrvs-beacon"
)

// FIXME: Make these configurable.
const (
	// HealthCheckInterval determines the frequency of regular healthchecks.
	HealthCheckInterval = 1 * time.Minute
	// HealthCheckTimeout
	HealthCheckTimeout = 3 * time.Second
	// FailedCheckThreshold is used to hide servers after failing healthchecks.
	FailedCheckThreshold = 15 // 15 minutes
	// PassedCheckThreshold is used to show servers again after being marked unhealthy.
	PassedCheckThreshold = 1
	// MaxFailedChecks is used to prune servers from the list entirely.
	MaxFailedChecks = 10080 // 7 days
)

func FilterHealthyServers(servers map[string]Server) map[string]Server {
	filtered := make(map[string]Server, 0)
	for k, s := range servers {
		if s.Healthy {
			filtered[k] = s
		}
	}
	return filtered
}

// SendHealthchecks queries the given servers and expires servers which are past
// the maximum number of failures.
func SendHealthchecks(servers map[string]Server) {
	for {
		time.Sleep(HealthCheckInterval)
		keysToDelete := make([]string, 0)
		for k, s := range servers {
			var ok bool
			servers[k], ok = UpdateServerHealth(s) // Overwrite self.
			if !ok {
				keysToDelete = append(keysToDelete, k)
			}
		}
		for _, k := range keysToDelete {
			// Log and remove from memory.
			// No need to store, since they will automatically register again.
			log.Println("removing unhealthy server after 7 days:", servers[k])
			delete(servers, k)
		}
	}
}

// UpdateServerHealth checks the health history of the given server, updating
// its state if necessary. Its second return value is false when the server
// should be deleted.
func UpdateServerHealth(s Server) (Server, bool) {
	// Mark servers unhealthy after three failed healthchecks.
	if !IsHealthy(s) {
		s.PassedChecks = 0
		s.FailedChecks++
		if s.FailedChecks >= FailedCheckThreshold {
			s.Healthy = false
		}
		if s.FailedChecks >= MaxFailedChecks {
			return s, false // should be deleted
		}
		return s, true // ok
	}

	// Healthcheck succeeded.
	s.PassedChecks++
	s.FailedChecks = 0

	// Mark unhealthy servers healthy again after three successful checks.
	if !s.Healthy && s.PassedChecks >= PassedCheckThreshold {
		s.Healthy = true
	}

	return s, true // ok
}

// IsHealthy sends a UDP beacon, and returns true when a healthcheck succeeds.
func IsHealthy(s Server) bool {
	_, err := beacon.GetServerReport(s.IP, s.Port+1000, HealthCheckTimeout)
	if err != nil {
		log.Printf("server %s:%d failed healthcheck: %v", s.IP, s.Port, err)
		return false
	}
	return true
}
