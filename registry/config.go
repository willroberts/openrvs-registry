package registry

import (
	"time"
)

type RegistryConfig struct {
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
