package registry

// ServerMap maps unique Hostport IDs to server metadata.
type ServerMap map[Hostport]Server

// Server contains all relevant fields for an individual game server.
type Server struct {
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
