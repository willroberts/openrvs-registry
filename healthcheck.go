// This file contains code which should live in the repo root and not this cmd.
// It's a bit of a dumping ground right now until I refactor.
package registry

import (
	"log"
	"sync"
	"time"

	beacon "github.com/ijemafe/openrvs-beacon"
)

// FIXME: Make these configurable.
const (
	// HealthCheckTimeout is the amount of time to wait before closing the UDP
	// socket.
	HealthCheckTimeout = 5 * time.Second // Values below 3 lose data.
	// FailedCheckThreshold is used to hide servers after failing healthchecks.
	FailedCheckThreshold = 60 // 30 minutes
	// PassedCheckThreshold is used to show servers again after being marked unhealthy.
	PassedCheckThreshold = 1
	// MaxFailedChecks is used to prune servers from the list entirely.
	MaxFailedChecks = 5760 // 2 days
)

// SendHealthchecks queries the given servers and filters servers which are past
// the maximum number of failures.
func SendHealthchecks(servers map[string]Server) map[string]Server {
	checked := make(map[string]Server, 0)

	var wg sync.WaitGroup
	var lock = sync.RWMutex{}

	for k, s := range servers {
		go func(k string, s Server) {
			wg.Add(1)
			var ok bool
			if s, ok = UpdateServerHealth(s); ok {
				lock.Lock()
				checked[k] = s
				lock.Unlock()
			} else {
				log.Println("removing unhealthy server after reaching maximum failure count:", servers[k])
			}
			wg.Done()
		}(k, s)
	}
	wg.Wait()

	log.Printf("out of %d servers, %d were healthy", len(servers),
		len(FilterHealthyServers(checked)))

	return checked
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
	if _, err := beacon.GetServerReport(s.IP, s.Port+1000, HealthCheckTimeout); err != nil {
		return false
	}

	return true
}

func FilterHealthyServers(servers map[string]Server) map[string]Server {
	filtered := make(map[string]Server, 0)
	for k, s := range servers {
		if s.Healthy {
			filtered[k] = s
		}
	}
	return filtered
}
