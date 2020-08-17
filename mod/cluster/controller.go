package cluster

// Cluster Controller

import (
	"log"

	"gitlab.com/infra.run/public/b3scale/mod/config"
)

// A Controller manages the back- and frontends.
// It creates instances based on the config source.
type Controller struct {
	backends  []*Backend
	frontends []*Frontend

	backendsConfig  config.BackendsConfig
	frontendsConfig config.FrontendsConfig
}

// NewController creates a new cluster controller
// instance with a config source.
func NewController(
	backendsConfig config.BackendsConfig,
	frontendsConfig config.FrontendsConfig,
) *Controller {
	return &Controller{
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
	_ = configs
}

// LogStatus collects cluster information and writes
// them to the log.
func (c *Controller) LogStatus() {
	log.Println("Cluster controller status...")
}
