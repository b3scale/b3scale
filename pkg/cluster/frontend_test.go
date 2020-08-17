package cluster

import (
	"testing"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

// TestNewFrontend by creating a frontend from a config
func TestNewFrontend(t *testing.T) {
	f := NewFrontend(&config.Frontend{
		Key:    "bigbadbutton",
		Secret: "seecret",
	})

	if f.ID != "bigbadbutton" {
		t.Error("Unexpected ID:", f.ID)
	}
}
