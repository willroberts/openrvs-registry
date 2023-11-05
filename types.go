package registry

import "fmt"

// Hostport represents an IP+Port combo to be used as a unique server ID.
type Hostport string

// NewHostport combines an IP and Port into a unique server ID.
func NewHostport(ip string, port int) Hostport {
	return Hostport(fmt.Sprintf("%s:%d", ip, port))
}

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

// GameTypes contains a map of all active game types, mapping them to either
// Adversarial or Cooperative mode identifiers 'adv' and 'coop'.
var GameTypes = map[string]string{
	// Raven Shield modes
	"RGM_BombAdvMode":           "adv",  // Bomb
	"RGM_DeathmatchMode":        "adv",  // Survival
	"RGM_EscortAdvMode":         "adv",  // Pilot
	"RGM_HostageRescueAdvMode":  "adv",  // Hostage
	"RGM_HostageRescueCoopMode": "coop", // Hostage Rescue
	"RGM_HostageRescueMode":     "coop",
	"RGM_MissionMode":           "coop", // Mission
	"RGM_SquadDeathmatch":       "adv",
	"RGM_SquadTeamDeathmatch":   "adv",
	"RGM_TeamDeathmatchMode":    "adv",  // Team Survival
	"RGM_TerroristHuntCoopMode": "coop", // Terrorist Hunt
	"RGM_TerroristHuntMode":     "coop",

	// Athena Sword modes
	"RGM_CaptureTheEnemyAdvMode": "adv",
	"RGM_CountDownMode":          "coop",
	"RGM_KamikazeMode":           "adv",
	"RGM_ScatteredHuntAdvMode":   "adv",
	"RGM_TerroristHuntAdvMode":   "adv",

	// TODO: Add Iron Wrath modes
	// Free Backup, Gas Alert, Intruder, Limited Seats, Virus Upload (all adv)
}
