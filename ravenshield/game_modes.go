package ravenshield

// GameModes contains a map of all active game types, mapping them to either
// Adversarial or Cooperative mode identifiers 'adv' and 'coop'.
var GameModes = map[string]string{
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
