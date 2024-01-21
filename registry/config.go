package registry

import (
	"time"
)

// Config contains the configuration values for the Registry service.
type Config struct {
	SeedPath           string
	CheckpointPath     string
	CheckpointInterval time.Duration

	HealthcheckInterval           time.Duration
	HealthcheckTimeout            time.Duration
	HealthcheckHealthyThreshold   int
	HealthcheckUnhealthyThreshold int
	HealthcheckHiddenThreshold    int

	ListenAddr string
}
