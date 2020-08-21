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

func TestGetBackendByID(t *testing.T) {
	backends := []*Backend{
		NewBackend(&config.Backend{Host: "host1"}),
		NewBackend(&config.Backend{Host: "host2"}),
	}
	ctrl := &Controller{backends: backends}
	be := ctrl.GetBackendByID("host1")
	if be == nil {
		t.Error("Expected host1 to be present")
	}

	be = ctrl.GetBackendByID("tsoh1")
	if be != nil {
		t.Error("Host tsoh1 should not be present")
	}
}

func TestRemoveBackend(t *testing.T) {
	backend := NewBackend(&config.Backend{Host: "host2"})
	backends := []*Backend{
		NewBackend(&config.Backend{Host: "host1"}),
		backend,
	}
	ctrl := &Controller{backends: backends}
	ctrl.RemoveBackend(backend)

	if len(ctrl.backends) != 1 {
		t.Error("Expected 1 backend.")
	}
}

func TestAddBackend(t *testing.T) {
	ctrl := &Controller{backends: []*Backend{}}

	backend1 := NewBackend(&config.Backend{
		Host:   "host1",
		Secret: "secret1",
	})
	backend1a := NewBackend(&config.Backend{
		Host:   "host1",
		Secret: "secret1",
	})

	// Let's just add the host
	ctrl.AddBackend(backend1)

	// This should do nothing:
	ctrl.AddBackend(backend1a)
	if ctrl.backends[0] != backend1 {
		t.Error("Expected operation to be idempotent.")
	}

	// This should replace the host
	backend1b := NewBackend(&config.Backend{
		Host:   "host1",
		Secret: "secret2",
	})
	ctrl.AddBackend(backend1b)
	if ctrl.backends[0] != backend1b {
		t.Error("Expected new cluster node.")
	}

}
