package v2

import v1 "github.com/willroberts/openrvs-registry"

type Registry interface {
}

type registry struct {
	Config RegistryConfig
	CSV    v1.CSVSerializer
}

func NewRegistry(config RegistryConfig) Registry {
	return &registry{
		Config: config,
		CSV:    v1.NewCSVSerializer(),
	}
}
