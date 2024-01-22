// Package registry provides automated management of the OpenRVS server list.
package registry

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	beacon "github.com/willroberts/openrvs-beacon"
	"github.com/willroberts/openrvs-registry/ravenshield"
)

// Registry maintains a list of servers, with functionality for healthchecking
// and loading/saving servers in CSV format.
type Registry interface {
	LoadServers(csvFile string) error
	SaveServers(csvFile string) error
	AddServer(ip string, data []byte) error
	ServerCount() int
	SendHealthchecks(onHealthy func(GameServer), onUnhealthy func(GameServer))

	HandleHTTP(listenAddress string) error
	HandleUDP(port int, h UDPHandler, stopCh chan struct{}) error
}

type registry struct {
	Config            Config
	CSV               CSVSerializer
	GameServerMap     GameServerMap
	GameServerMapLock sync.RWMutex
}

// NewRegistry initializes and returns a Registry.
func NewRegistry(config Config) Registry {
	return &registry{
		Config:        config,
		CSV:           NewCSVSerializer(),
		GameServerMap: make(GameServerMap),
	}
}

func (r *registry) LoadServers(csvFile string) error {
	b, err := os.ReadFile(csvFile)
	if err != nil {
		return err
	}

	parsed, err := r.CSV.Deserialize(b)
	if err != nil {
		return err
	}
	r.GameServerMapLock.Lock()
	r.GameServerMap = parsed
	r.GameServerMapLock.Unlock()

	return nil
}

func (r *registry) SaveServers(csvFile string) error {
	data := r.CSV.Serialize(r.GameServerMap)
	return os.WriteFile(r.Config.CheckpointPath, data, 0644)
}

func (r *registry) AddServer(ip string, data []byte) error {
	if net.ParseIP(ip).IsPrivate() {
		return errors.New("skipping server with private IP")
	}

	report, err := beacon.ParseServerReport(ip, data)
	if err != nil {
		return err
	}

	if report.ServerName == "" {
		return errors.New("skipping server with no name")
	}

	if report.Port == 0 {
		return errors.New("skipping server with no port")
	}

	if report.CurrentMode == "" {
		return errors.New("skipping server with no game mode")
	}

	// Manually healthcheck this server before adding it to the map.
	serverID := fmt.Sprintf("%s:%d", report.IPAddress, report.Port)
	server = r.updateServerHealth(GameServer{
		Name:       report.ServerName,
		IP:         report.IPAddress,
		Port:       report.Port,
		BeaconPort: report.BeaconPort,
		GameMode:   ravenshield.GameModes[report.CurrentMode],
	}, func(GameServer) {}, func(GameServer) {})

	r.GameServerMapLock.Lock()
	r.GameServerMap[serverID] = server
	r.GameServerMapLock.Unlock()

	return nil
}

func (r *registry) ServerCount() int {
	return len(r.GameServerMap)
}

func (r *registry) SendHealthchecks(
	onHealthy func(s GameServer),
	onUnhealthy func(s GameServer),
) {
	var (
		output = make(GameServerMap)
		wg     sync.WaitGroup
		lock   sync.RWMutex
	)

	for hostport, server := range r.GameServerMap {
		wg.Add(1)
		go func(hostport string, server GameServer) {
			lock.Lock()
			defer lock.Unlock()
			output[hostport] = r.updateServerHealth(server, onHealthy, onUnhealthy)
			wg.Done()
		}(hostport, server)
	}
	wg.Wait()

	r.GameServerMapLock.Lock()
	r.GameServerMap = output
	r.GameServerMapLock.Unlock()
}

func (r *registry) updateServerHealth(
	s GameServer,
	onHealthy func(GameServer),
	onUnhealthy func(GameServer),
) GameServer {
	if s.BeaconPort == 0 {
		// Assume BeaconPort is Port+1000 if we don't know it yet.
		s.BeaconPort = s.Port + 1000
	}
	reportBytes, err := beacon.GetServerReport(s.IP, s.BeaconPort, r.Config.HealthcheckTimeout)
	if err != nil {
		s.Health.PassedChecks = 0 // 0 checks in a row have passed
		s.Health.FailedChecks++   // Another check in a row has failed
		if s.Health.FailedChecks == r.Config.HealthcheckUnhealthyThreshold {
			onUnhealthy(s)
			s.Health.Healthy = false // Too many failed checks in a row.
		}
		if s.Health.FailedChecks >= r.Config.HealthcheckHiddenThreshold {
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
		s.GameMode = ravenshield.GameModes[report.CurrentMode]
	}

	// Mark unhealthy servers healthy again after consecutive successful checks.
	if !s.Health.Healthy && s.Health.PassedChecks >= r.Config.HealthcheckHealthyThreshold {
		s.Health.Healthy = true // Server is healthy again.
		onHealthy(s)
	}

	return s
}
