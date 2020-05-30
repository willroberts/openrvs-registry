package registry

import "fmt"

type Server struct {
	// Fields exposed to clients.
	Name     string
	IP       string
	Port     int
	GameMode string

	// Internal fields.
	healthy      bool
	passedChecks int
	failedChecks int
}

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

func HostportToKey(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
