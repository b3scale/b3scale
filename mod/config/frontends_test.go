package config

import (
	"testing"
)

// TestGetFrontends is getting all frontends from the example config
func TestGetFrontends(t *testing.T) {
	c := NewFrontendsConfig("test/data/config/frontends.conf")

	f, err := c.GetFrontends()
	if err != nil {
		t.Error(err)
	}

	if len(f) != 1 {
		t.Error("Expected 1 frontend. Got:", len(f))
	}

}
