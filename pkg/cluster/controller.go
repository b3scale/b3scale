package cluster

import (
	"log"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

// The Controller manages the state, reads configs
// creates instances and adds them to the cluster state
type Controller struct {
	state           *State
	backendsConfig  config.BackendsConfig
	frontendsConfig config.FrontendsConfig
}

// NewController creates a new cluster controller instance.
func NewController(
	state *State,
	backendsConfig config.BackendsConfig,
	frontendsConfig config.FrontendsConfig,
) *Controller {
	return &Controller{
		state:           state,
		backendsConfig:  backendsConfig,
		frontendsConfig: frontendsConfig,
	}
}

// Start the controller process
func (ctrl *Controller) Start() {
	log.Println("Starting cluster controller.")
	err := ctrl.Reload()
	if err != nil {
		panic(err) // Initial boot must succeed.
	}

	ctrl.state.Log()
}

// Reload the entire cluster
func (ctrl *Controller) Reload() error {
	log.Println("Reloading cluster configuration.")
	if err := ctrl.ReloadBackends(); err != nil {
		return err
	}
	if err := ctrl.ReloadFrontends(); err != nil {
		return err
	}
	return nil
}

// ReloadBackends refreshes the backend configurations
func (ctrl *Controller) ReloadBackends() error {
	// Get configured backends
	configs, err := ctrl.backendsConfig.Load()
	if err != nil {
		return err
	}

	backends := make([]*Backend, 0, len(configs))
	for _, cfg := range configs {
		backends = append(backends, NewBackend(cfg))
	}
	ctrl.state.UpdateBackends(backends)
	return nil
}

// ReloadFrontends refreshes the frontend configurations
func (ctrl *Controller) ReloadFrontends() error {
	// Get configured backends
	configs, err := ctrl.frontendsConfig.Load()
	if err != nil {
		return err
	}

	frontends := make([]*Frontend, 0, len(configs))
	for _, cfg := range configs {
		frontends = append(frontends, NewFrontend(cfg))
	}
	ctrl.state.UpdateFrontends(frontends)
	return nil
}
