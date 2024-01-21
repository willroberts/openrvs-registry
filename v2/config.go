package v2

import (
	"time"

	v1 "github.com/willroberts/openrvs-registry"
)

type RegistryConfig struct {
	SeedPath            string
	CheckpointPath      string
	CheckpointInterval  time.Duration
	HealthcheckInterval time.Duration
	HealthcheckTimeout  time.Duration
	ListenAddr          v1.Hostport
}
