package v2

import (
	"sync"
	"time"

	beacon "github.com/willroberts/openrvs-beacon"
)

const (
	// failedCheckThreshold is used to hide servers after failing healthchecks.
	failedCheckThreshold = 60 // 30 minutes

	// passedCheckThreshold is used to show servers again after being marked unhealthy.
	passedCheckThreshold = 1

	// maxFailedChecks is used to prune servers from the list entirely.
	maxFailedChecks = 5760 // 2 days
)

// SendHealthchecks queries each known server and updates its health status in
// memory.
// TODO: Automatically prune servers which have failed healthchecks for a week.
// This would be a bit easier if we stored the "last passed check" timestamp in
// the CSV file, since restarting the service currently restarts the number of
// consecutive failed checks (and we need over 20,000 failed checks to
// constitute a week).
func SendHealthchecks(
	servers GameServerMap,
	timeout time.Duration,
	onHealthy func(GameServer),
	onUnhealthy func(GameServer),
) GameServerMap {
	var (
		checked = make(GameServerMap, 0) // Output map.
		wg      sync.WaitGroup           // For synchronizing the UDP beacons.
		lock    = sync.RWMutex{}         // For safely accessing checked map.
	)

	for hostport, s := range servers {
		// Kick off this work in a new thread.
		wg.Add(1)
		go func(hostport string, s GameServer) {
			updated := updateHealth(s, timeout, onHealthy, onUnhealthy)
			lock.Lock()
			checked[hostport] = updated
			lock.Unlock()
			wg.Done()
		}(hostport, s)
	}
	wg.Wait()

	return checked
}

// updateHealth modifies and returns a GameServer based on a healthcheck
// result.
func updateHealth(
	s GameServer,
	timeout time.Duration,
	onHealthy func(GameServer),
	onUnhealthy func(GameServer),
) GameServer {
	// Send a UDP beacon and determine if it failed.
	var failed bool
	reportBytes, err := beacon.GetServerReport(s.IP, s.Port+1000, timeout)
	if err != nil {
		failed = true // No need to log connection refused, timeout, etc.
	}

	// Healthcheck failed.
	if failed {
		s.Health.PassedChecks = 0 // 0 checks in a row have passed
		s.Health.FailedChecks++   // Another check in a row has failed
		if s.Health.FailedChecks == failedCheckThreshold {
			onUnhealthy(s)
			s.Health.Healthy = false // Too many failed checks in a row.
		}
		if s.Health.FailedChecks >= maxFailedChecks {
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
		s.Health.ParseFailed = true
	} else {
		s.Health.ParseFailed = false
		s.Name = report.ServerName
		s.GameMode = report.CurrentMode
	}

	// Mark unhealthy servers healthy again after three successful checks.
	if !s.Health.Healthy && s.Health.PassedChecks >= passedCheckThreshold {
		s.Health.Healthy = true // Server is healthy again.
		onHealthy(s)
	}

	return s
}

// FilterHealthyServers iterates through a map of Servers, and returns a subset
// maps of Servers which only contains the servers marked as being healthy.
func FilterHealthyServers(servers GameServerMap) GameServerMap {
	filtered := make(GameServerMap)
	for k, s := range servers {
		if s.Health.Healthy {
			filtered[k] = s // Copy to output map.
		}
	}
	return filtered
}
