package registry

// GameServerMap maps unique Hostport IDs to server metadata.
type GameServerMap map[Hostport]GameServer

// GameServer contains all relevant fields for an individual game server.
type GameServer struct {
	Name     string
	IP       string
	Port     int
	GameMode string

	Health HealthStatus
}

// HealthStatus contains information needed to track whether a server is healthy.
type HealthStatus struct {
	Healthy      bool
	Expired      bool
	PassedChecks int
	FailedChecks int
}
