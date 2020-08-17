package config

// Backend is the configuration of a cluster node
type Backend struct {
	Host   string
	Secret string
}

// Frontend is the configuration of a
// a cluster consumer.
type Frontend struct {
	Key    string
	Secret string
}

// BackendsConfig interface
type BackendsConfig interface {
	Load() ([]*Backend, error)
}

// FrontendsConfig interface
type FrontendsConfig interface {
	Load() ([]*Frontend, error)
}
