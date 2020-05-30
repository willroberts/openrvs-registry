// This file contains code which should live in the repo root and not this cmd.
// It's a bit of a dumping ground right now until I refactor.
package registry

import (
	"log"
	"time"

	beacon "github.com/ijemafe/openrvs-beacon"
)

const (
	// HealthCheckInterval determines the frequency of regular healthchecks.
	HealthCheckInterval = 1 * time.Minute
	// FailedCheckThreshold is used to hide servers after failing healthchecks.
	FailedCheckThreshold = 15
	// PassedCheckThreshold is used to show servers again after being marked unhealthy.
	PassedCheckThreshold = 2
	// MaxFailedChecks is used to prune servers from the list entirely.
	MaxFailedChecks = 10080
)

// Healthcheck sends a UDP beacon, and returns true when a healthcheck succeeds.
func Healthcheck(s Server) bool {
	_, err := beacon.GetServerReport(s.IP, s.Port+1000)
	if err != nil {
		log.Println("healthcheck err:", err)
		return false
	}
	return true
}

// UpdateServerHealth contains logic for updating the server map.
func UpdateServerHealth(s Server, succeeded bool) {
	// Mark servers unhealthy after three failed healthchecks.
	if !succeeded {
		s.passedChecks = 0
		s.failedChecks++
		if s.failedChecks >= FailedCheckThreshold {
			s.healthy = false
		}
		if s.failedChecks >= MaxFailedChecks {
			// TODO: Take action here.
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
