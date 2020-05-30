// This file contains code which should live in the repo root and not this cmd.
// It's a bit of a dumping ground right now until I refactor.
package registry

import (
	"log"
	"time"

	beacon "github.com/ijemafe/openrvs-beacon"
)

const (
	HealthCheckInterval  = 1 * time.Minute
	FailedCheckThreshold = 15    // Hide servers after being down 15 mins.
	PassedCheckThreshold = 2     // Show servers again after passing 2 checks.
	MaxFailedChecks      = 10080 // Prune servers from the list entirely after being down 7 days.
)

func healthcheck(s Server) {
	var failed bool
	_, err := beacon.GetServerReport(s.IP, s.Port+1000)
	if err != nil {
		log.Println("healthcheck err:", err)
		failed = true
	}

	// Mark servers unhealthy after three failed healthchecks.
	if failed {
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
