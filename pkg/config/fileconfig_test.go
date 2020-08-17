package config

import (
	"testing"
)

// TestFrontendsFileConfig is getting all frontends from the example config
func TestFrontendsFileConfig(t *testing.T) {
	c := NewFrontendsFileConfig("../../test/data/config/frontends.conf")

	frontends, err := c.Load()
	if err != nil {
		t.Error(err)
	}

	if len(frontends) != 1 {
		t.Error("Expected 1 frontend. Got:", len(frontends))
	}

	f := frontends[0]
	if f.Key != "bigbluebutton" && f.Secret != "v3rYSt0Ng S3Cr3th" {
		t.Error("Unexpected key and secret.", f)
	}
}

// TestBackendsFileConfig is getting all frontends from the example config
func TestBackendsFileConfig(t *testing.T) {
	c := NewBackendsFileConfig("../../test/data/config/nodes.conf")

	backends, err := c.Load()
	if err != nil {
		t.Error(err)
	}

	if len(backends) != 2 {
		t.Error("Expected 2 backends. Got:", len(backends))
	}

	b := backends[0]
	if b.Host != "https://fooo.bar/api/" {
		t.Error("Expected host to be: https://fooo.bar/api. Got:", b.Host)
	}
}
