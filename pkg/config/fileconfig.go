package config

// FileConfigs read configuration from files.

import (
	"io/ioutil"
	"strings"
)

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
	// Read file and parse contents
	data, err := ioutil.ReadFile(c.File)
	if err != nil {
		return nil, err
	}

	// Results
	frontends := []*Frontend{}

	config := parseConfig(data)
	// Decode config and create frontend configs
	// from the data.
	for _, cmd := range config {
		if len(cmd) != 3 {
			continue
		}
		if cmd[0] != "frontend" {
			continue
		}

		// Make frontend config and append to configs
		frontends = append(frontends, NewFrontend(cmd[1], cmd[2]))
	}

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
	// Results
	backends := []*Backend{}

	// Read file and parse contents
	data, err := ioutil.ReadFile(c.File)
	if err != nil {
		return nil, err
	}

	// Decode config and create frontend configs
	// from the data.
	config := parseConfig(data)
	for _, cmd := range config {
		if len(cmd) != 3 {
			continue
		}
		if cmd[0] != "node" {
			continue
		}
		// Make backend config and append to configs
		backends = append(backends, NewBackend(cmd[1], cmd[2]))

	}
	return backends, nil
}

// File parseing helper. Just splits the config
// lines into tokens.
func parseConfig(data []byte) [][]string {
	config := [][]string{}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		cmd := strings.SplitN(line, " ", 3)
		config = append(config, cmd)
	}
	return config
}
