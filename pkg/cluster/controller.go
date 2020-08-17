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
func (c *Controller) Start() {
	log.Println("Starting cluster controller.")

	// Load configurations
	c.Reload()

	// Log initial status
	c.LogStatus()
}

// Reload all configurations
func (c *Controller) Reload() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.reloadBackends()
	c.reloadFrontends()
}

func (c *Controller) reloadBackends() {
	configs, err := c.backendsConfig.Load()
	if err != nil {
		// Log the error but keep on going.
		log.Println("Error while loading backends:", err)
		return
	}

	_ = configs
	// Update our backend instances: Create new
	// instances if the backend identified by it's
	// host URL is unknown.
	//
	// Remove node instances, no longer present
	// in the configuration.

}

func (c *Controller) reloadFrontends() {
	configs, err := c.frontendsConfig.Load()
	if err != nil {
		// Log the error but keep on going.
		log.Println("Error while loading frontends:", err)
		return
	}

	// Register all frontends from the config
	registeredIDs := []string{}
	for _, config := range configs {
		// Registering is idempotent
		f := NewFrontend(config)
		c.AddFrontend(f)
		registeredIDs = append(registeredIDs, f.ID)
	}

	// Remove all frontends not longer in the config
	for _, frontend := range c.frontends {
		present := false
		for _, id := range registeredIDs {
			if frontend.ID == id {
				present = true
				break
			}
		}
		if !present {
			log.Println("Unregistering frontend:", frontend.ID)
			c.removeFrontend(frontend)
		}
	}

}

// AddFrontend adds a frontend to the cluster.
// This is an idempotent operation. If the frontend id
// is already registered, it will be replaced with
// the new frontend.
func (c *Controller) AddFrontend(frontend *Frontend) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.addFrontend(frontend)
}

// Unsafe interal add frontend
func (c *Controller) addFrontend(frontend *Frontend) {
	c.removeFrontend(frontend)
	c.frontends = append(c.frontends, frontend)
	log.Println("Registered frontend:", frontend.config.Key)
}

// RemoveFrontend removes a frontend from the cluster
func (c *Controller) RemoveFrontend(frontend *Frontend) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.removeFrontend(frontend)
}

// Unsafe internal removeFrontend without locking
func (c *Controller) removeFrontend(frontend *Frontend) {
	frontends := make([]*Frontend, 0, len(c.frontends))
	for _, f := range c.frontends {
		if f.ID == frontend.ID {
			continue
		}
		frontends = append(frontends, f)
	}

	c.frontends = frontends
}

// GetFrontendByID retrievs a frontend identified by
// its key from our list of frontends.
func (c *Controller) GetFrontendByID(id string) *Frontend {
	for _, f := range c.frontends {
		if f.ID == id {
			return f
		}
	}
	return nil
}

// GetFrontends retrievs all frontends in the controller
func (c *Controller) GetFrontends() []*Frontend {
	return c.frontends
}

// LogStatus collects cluster information and writes
// them to the log.
func (c *Controller) LogStatus() {
	log.Println("Cluster controller status...")
}
