package v2

// GameServerMap maps unique server IDs to server metadata.
type GameServerMap map[string]GameServer

// GameServer contains all relevant fields for an individual game server.
type GameServer struct {
	Name     string
	IP       string
	Port     int
	GameMode string

	Health GameServerHealthStatus
}

// GameServerHealthStatus contains information needed to track whether a server
// is healthy.
type GameServerHealthStatus struct {
	Healthy      bool
	Expired      bool
	PassedChecks int
	FailedChecks int
	ParseFailed  bool
}
