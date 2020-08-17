package cluster

import (
	"testing"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

// TestAddFrontend by adding a frontend to the controller
func TestAddFrontend(t *testing.T) {
	c := NewController(nil, nil)
	f := NewFrontend(&config.Frontend{Key: "foo", Secret: "bar"})
	c.addFrontend(f)

	if len(c.frontends) != 1 {
		t.Error("Expected 1 frontend, got:", len(c.frontends))
	}
}
