package cluster

import (
	"testing"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

func TestAddFrontend(t *testing.T) {
	c := NewController(nil, nil)
	f1 := NewFrontend(&config.Frontend{Key: "foo", Secret: "bar"})
	f2 := NewFrontend(&config.Frontend{Key: "foo", Secret: "b4r"})

	c.AddFrontend(f1)
	if len(c.frontends) != 1 {
		t.Error("Expected 1 frontend, got:", len(c.frontends))
	}

	// Again
	c.AddFrontend(f2)
	if len(c.frontends) != 1 {
		t.Error("Expected 1 frontend, got:", len(c.frontends))
	}

	if c.frontends[0].config.Secret != "b4r" {
		t.Error("Expected a config update, got:", c.frontends[0].config)
	}
}

func TestGetFrontendByID(t *testing.T) {
	c := NewController(nil, nil)
	f1 := NewFrontend(&config.Frontend{Key: "foo", Secret: "bar"})
	f2 := NewFrontend(&config.Frontend{Key: "bar", Secret: "b4r"})
	c.AddFrontend(f1)
	c.AddFrontend(f2)

	f := c.GetFrontendByID("bar")
	if f == nil {
		t.Error("Expected bar to be a frontend.")
	}

	f = c.GetFrontendByID("asdf")
	if f != nil {
		t.Error("Did not expect asdf to be a frontend.")
	}

}
