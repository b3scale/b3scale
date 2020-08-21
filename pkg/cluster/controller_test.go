package cluster

import (
	"testing"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

type feConfStub struct {
	configs []*config.Frontend
}

func (f *feConfStub) Load() ([]*config.Frontend, error) {
	return f.configs, nil
}

type beConfStub struct {
	configs []*config.Backend
}

func (b *beConfStub) Load() ([]*config.Backend, error) {
	return b.configs, nil
}

func TestReload(t *testing.T) {
	state := NewState()

	backendConfigs := &beConfStub{[]*config.Backend{
		&config.Backend{Host: "host1", Secret: "secret1"},
	}}
	frontendConfigs := &feConfStub{[]*config.Frontend{
		&config.Frontend{Key: "fe1", Secret: "fesecret"},
	}}

	ctrl := NewController(
		state,
		backendConfigs,
		frontendConfigs,
	)
	ctrl.Start()

	if state.backends[0].ID != "host1" {
		t.Error("expected 1 backend: host1")
	}

	// Add frontend
	frontendConfigs.configs = append(
		frontendConfigs.configs,
		&config.Frontend{
			Key:    "fe2",
			Secret: "foo",
		})
	backendConfigs.configs = append(
		backendConfigs.configs,
		&config.Backend{
			Host:   "host2",
			Secret: "secret2",
		})

	ctrl.Reload()

	// Remove backend
	backendConfigs.configs = []*config.Backend{
		&config.Backend{Host: "host1", Secret: "secret1"},
	}
	ctrl.Reload()

	backendConfigs.configs = []*config.Backend{
		&config.Backend{Host: "host1", Secret: "secret2"},
	}
	ctrl.Reload()
}
