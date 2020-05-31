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

// SendHealthchecks ... ?
func SendHealthchecks(servers map[string]Server) map[string]Server {
	var (
		checked = make(map[string]Server, 0)
		wg      sync.WaitGroup
		lock    = sync.RWMutex{}
	)

	for k, s := range servers {
		wg.Add(1)
		go func(k string, s Server) {
			updated := UpdateHealthStatus(s)
			lock.Lock()
			checked[k] = updated // All servers are updated, healthy or not.
			lock.Unlock()
			wg.Done()
		}(k, s)
	}
	wg.Wait()

	log.Println("healthy servers:", len(FilterHealthyServers(checked)), "out of", len(servers))
	return checked
}

// UpdateHealthStatus modifies and returns an object according to a healthcheck
// result.
func UpdateHealthStatus(s Server) Server {
	var failed bool
	if _, err := beacon.GetServerReport(s.IP, s.Port+1000, HealthCheckTimeout); err != nil {
		failed = true // No need to log connection refused, timeout, etc.
	}

	if failed {
		s.Health.PassedChecks = 0
		s.Health.FailedChecks++
		if s.Health.FailedChecks == FailedCheckThreshold {
			log.Println("server is now unhealthy:", s.IP, s.Port)
			s.Health.Healthy = false
		}
		if s.Health.FailedChecks >= MaxFailedChecks {
			s.Health.Expired = true
		}
		return s
	}

	// Healthcheck succeeded.
	s.Health.PassedChecks++
	s.Health.FailedChecks = 0

	// Mark unhealthy servers healthy again after three successful checks.
	if !s.Health.Healthy && s.Health.PassedChecks >= PassedCheckThreshold {
		s.Health.Healthy = true
		log.Println("server is now healthy:", s.IP, s.Port)
	}

	return s
}

func FilterHealthyServers(servers map[string]Server) map[string]Server {
	filtered := make(map[string]Server, 0)
	for k, s := range servers {
		if s.Health.Healthy {
			filtered[k] = s
		}
	}
	return filtered
}
