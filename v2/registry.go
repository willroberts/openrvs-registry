package v2

// This class is not yet used.
type Registry interface {
}

type registry struct {
	Config RegistryConfig
}

func NewRegistry(config RegistryConfig) Registry {
	return &registry{
		Config: config,
	}
}
