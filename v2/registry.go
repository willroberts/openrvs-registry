package v2

import v1 "github.com/willroberts/openrvs-registry"

type Registry interface {
}

type registry struct {
	Config        RegistryConfig
	CSV           v1.CSVSerializer
	GameServerMap v1.GameServerMap
}

func NewRegistry(config RegistryConfig) Registry {
	return &registry{
		Config: config,
		CSV:    v1.NewCSVSerializer(),
	}
}

func (r *registry) AddServer(ip string, data []byte) {
	// Not yet implemented.
}
