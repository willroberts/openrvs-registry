package v2

import "time"

type RegistryConfig struct {
	SeedPath            string
	CheckpointPath      string
	CheckpointInterval  time.Duration
	HealthcheckInterval time.Duration
	IgnoredNetworks     []string
}
