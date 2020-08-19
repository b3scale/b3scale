package cluster

// Cluster Controller

import (
	"log"
	"sync"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

// A Controller manages the back- and frontends.
// It creates instances based on the config source.
type Controller struct {
	backends  []*Backend
	frontends []*Frontend

	backendsConfig  config.BackendsConfig
	frontendsConfig config.FrontendsConfig

	mtx sync.Mutex
}

// NewController creates a new cluster controller
// instance with a config source.
func NewController(
	backendsConfig config.BackendsConfig,
	frontendsConfig config.FrontendsConfig,
) *Controller {
	return &Controller{
		frontends:       []*Frontend{},
		backends:        []*Backend{},
		backendsConfig:  backendsConfig,
		frontendsConfig: frontendsConfig,
	}
}

// Start the cluster
func (ctrl *Controller) Start() {
	log.Println("Starting cluster controller.")

	// Load configurations
	ctrl.Reload()

	// Log initial status
	ctrl.LogStatus()
}

// Reload all configurations
func (ctrl *Controller) Reload() {
	ctrl.mtx.Lock()
	defer ctrl.mtx.Unlock()

	ctrl.reloadBackends()
	ctrl.reloadFrontends()
}

func (ctrl *Controller) reloadBackends() {
	configs, err := ctrl.backendsConfig.Load()
	if err != nil {
		// Log the error but keep on going.
		log.Println("Error while loading backends:", err)
		return
	}

	// Add all instances to our cluster.
	registeredBackends := []string{}
	for _, c := range configs {
		// Create new backend instance
		b := NewBackend(c)
		ctrl.addBackend(b)
		registeredBackends = append(registeredBackends, c.Host)
	}

	//
	// Remove node instances, no longer present
	// in the configuration.
}

// GetBackendByID retrieves a backend by it's ID
func (ctrl *Controller) GetBackendByID(id string) *Backend {
	for _, b := range ctrl.backends {
		if b.ID == id {
			return b
		}
	}
	return nil
}

func (ctrl *Controller) addBackend(backend *Backend) {
	current := ctrl.GetBackendByID(backend.ID)
}

func (ctrl *Controller) reloadFrontends() {
	configs, err := ctrl.frontendsConfig.Load()
	if err != nil {
		// Log the error but keep on going.
		log.Println("Error while loading frontends:", err)
		return
	}

	// Register all frontends from the config
	registeredIDs := []string{}
	for _, c := range configs {
		// Registering is idempotent
		f := NewFrontend(c)
		ctrl.addFrontend(f)
		registeredIDs = append(registeredIDs, f.ID)
	}

	// Remove all frontends not longer in the config
	for _, frontend := range ctrl.frontends {
		present := false
		for _, id := range registeredIDs {
			if frontend.ID == id {
				present = true
				break
			}
		}
		if !present {
			log.Println("Unregistering frontend:", frontend.ID)
			ctrl.removeFrontend(frontend)
		}
	}

}

// AddFrontend adds a frontend to the cluster.
// This is an idempotent operation. If the frontend id
// is already registered, it will be replaced with
// the new frontend.
func (ctrl *Controller) AddFrontend(frontend *Frontend) {
	ctrl.mtx.Lock()
	defer ctrl.mtx.Unlock()
	ctrl.addFrontend(frontend)
}

// Unsafe interal add frontend
func (ctrl *Controller) addFrontend(frontend *Frontend) {
	ctrl.removeFrontend(frontend)
	ctrl.frontends = append(ctrl.frontends, frontend)
	log.Println("Registered frontend:", frontend.config.Key)
}

// RemoveFrontend removes a frontend from the cluster
func (ctrl *Controller) RemoveFrontend(frontend *Frontend) {
	ctrl.mtx.Lock()
	defer ctrl.mtx.Unlock()
	ctrl.removeFrontend(frontend)
}

// Unsafe internal removeFrontend without locking
func (ctrl *Controller) removeFrontend(frontend *Frontend) {
	frontends := make([]*Frontend, 0, len(ctrl.frontends))
	for _, f := range ctrl.frontends {
		if f.ID == frontend.ID {
			continue
		}
		frontends = append(frontends, f)
	}

	ctrl.frontends = frontends
}

// GetFrontendByID retrievs a frontend identified by
// its key from our list of frontends.
func (ctrl *Controller) GetFrontendByID(id string) *Frontend {
	for _, f := range ctrl.frontends {
		if f.ID == id {
			return f
		}
	}
	return nil
}

// GetFrontends retrievs all frontends in the controller
func (ctrl *Controller) GetFrontends() []*Frontend {
	return ctrl.frontends
}

// LogStatus collects cluster information and writes
// them to the log.
func (ctrl *Controller) LogStatus() {
	log.Println("Cluster controller status...")
}
