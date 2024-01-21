package v2

import (
	"errors"
	"net"
	"os"
	"sync"

	beacon "github.com/willroberts/openrvs-beacon"
	v1 "github.com/willroberts/openrvs-registry"
)

type Registry interface {
	GetConfig() RegistryConfig

	LoadServers(csvFile string) error
	SaveServers(csvFile string) error
	AddServer(ip string, data []byte) error
	ServerCount() int
	UpdateServerHealth(onHealthy func(v1.GameServer), onUnhealthy func(v1.GameServer))

	HandleHTTP(listenAddress v1.Hostport) error
}

type registry struct {
	Config            RegistryConfig
	CSV               v1.CSVSerializer
	GameServerMap     v1.GameServerMap
	GameServerMapLock sync.RWMutex
}

func NewRegistry(config RegistryConfig) Registry {
	return &registry{
		Config:        config,
		CSV:           v1.NewCSVSerializer(),
		GameServerMap: make(v1.GameServerMap),
	}
}

func (r *registry) GetConfig() RegistryConfig {
	return r.Config
}

func (r *registry) LoadServers(csvFile string) error {
	r.GameServerMapLock.Lock()
	defer r.GameServerMapLock.Unlock()

	b, err := os.ReadFile(csvFile)
	if err != nil {
		return err
	}

	parsed, err := r.CSV.Deserialize(b)
	if err != nil {
		return err
	}
	r.GameServerMap = parsed

	return nil
}

func (r *registry) SaveServers(csvFile string) error {
	data := r.CSV.Serialize(r.GameServerMap)
	return os.WriteFile(r.Config.CheckpointPath, data, 0644)
}

func (r *registry) AddServer(ip string, data []byte) error {
	r.GameServerMapLock.Lock()
	defer r.GameServerMapLock.Unlock()

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

	serverID := v1.NewHostport(report.IPAddress, report.Port)
	r.GameServerMap[serverID] = v1.GameServer{
		Name:     report.ServerName,
		IP:       report.IPAddress,
		Port:     report.Port,
		GameMode: v1.GameModes[report.CurrentMode],
	}

	return nil
}

func (r *registry) ServerCount() int {
	return len(r.GameServerMap)
}

func (r *registry) UpdateServerHealth(
	onHealthy func(s v1.GameServer),
	onUnhealthy func(s v1.GameServer),
) {
	r.GameServerMapLock.Lock()
	defer r.GameServerMapLock.Unlock()

	r.GameServerMap = v1.SendHealthchecks(
		r.GameServerMap,
		r.Config.HealthcheckTimeout,
		onHealthy,
		onUnhealthy,
	)
}
