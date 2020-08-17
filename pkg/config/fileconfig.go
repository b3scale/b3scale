package config

// FileConfigs read configuration from files.

// A FrontendsFileConfig reads frontends from
// a file.
type FrontendsFileConfig struct {
	File string
}

// NewFrontendsFileConfig creates a new FrontendsConfig
func NewFrontendsFileConfig(file string) *FrontendsFileConfig {
	c := &FrontendsFileConfig{
		File: file,
	}

	return c
}

// Load implements the Frontends interface
// and retrieves a list of cluster frontends
func (c *FrontendsFileConfig) Load() ([]*Frontend, error) {
	frontends := []*Frontend{}
	return frontends, nil
}

// A BackendsFileConfig reads backend configurations
// from a config file source
type BackendsFileConfig struct {
	File string
}

// NewBackendsFileConfig creates a new BackendsConfig
func NewBackendsFileConfig(file string) *BackendsFileConfig {
	c := &BackendsFileConfig{
		File: file,
	}

	return c
}

// Load implements the Backends interface and
// retrieves a list of cluster backend configurations.
func (c *BackendsFileConfig) Load() ([]*Backend, error) {
	backends := []*Backend{}
	return backends, nil
}
