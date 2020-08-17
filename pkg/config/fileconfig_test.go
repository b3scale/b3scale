package config

import (
	"testing"
)

// TestFrontendsFileConfig is getting all frontends from the example config
func TestFrontendsFileConfig(t *testing.T) {
	c := NewFrontendsFileConfig("test/data/config/frontends.conf")

	f, err := c.Load()
	if err != nil {
		t.Error(err)
	}

	if len(f) != 1 {
		t.Error("Expected 1 frontend. Got:", len(f))
	}

}

// TestBackendsFileConfig is getting all frontends from the example config
func TestBackendsFileConfig(t *testing.T) {
	c := NewBackendsFileConfig("test/data/config/backends.conf")

	f, err := c.Load()
	if err != nil {
		t.Error(err)
	}

	if len(f) != 1 {
		t.Error("Expected 2 backends. Got:", len(f))
	}
}
