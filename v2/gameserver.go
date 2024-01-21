package v2

type GameServerMap map[Hostport]*GameServer

type GameServer struct {
	Name       string
	IP         string
	Port       int
	BeaconPort int
	GameMode   string
	Health     HealthStatus
}

type HealthStatus struct {
	Healthy      bool
	Expired      bool
	PassedChecks int
	FailedChecks int
}
