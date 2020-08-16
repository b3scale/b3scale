package config

import (
	"gitlab.com/infra.run/public/b3scale/cluster"
)

// A FrontendsConfig points to a config file
// from which frontend configurations are read.
type FrontendsConfig struct {
	File string
}

// NewFrontendsConfig creates a new FrontendsConfig
func NewFrontendsConfig(file string) *FrontendsConfig {
	c := &FrontendsConfig{
		File: file,
	}

	return c
}

// GetFrontends retrieves a list of cluster frontends
func (c *FrontendsConfig) GetFrontends() ([]*cluster.Frontend, error) {
	frontends := []*cluster.Frontend{}
	return frontends, nil
}
