package cluster

// Cluster State

import (
	"log"
	"sync"
)

// The State of the cluster is a set of frontends
// and backends. Each representing the state of
// a single frontend / backend in the cluster.
type State struct {
	backends  []*Backend
	frontends []*Frontend

	mtx sync.Mutex
}

// NewState creates a new cluster controller
// instance with a config source.
func NewState() *State {
	return &State{
		frontends: []*Frontend{},
		backends:  []*Backend{},
	}
}

// Start the cluster
func (state *State) Start() {
	log.Println("Booting cluster.")
}

// UpdateBackends all configurations
func (state *State) UpdateBackends(backends []*Backend) {
	state.mtx.Lock()
	defer state.mtx.Unlock()
	state.updateBackends(backends)
}

func (state *State) updateBackends(backends []*Backend) {
	// Add all instances to our cluster.
	registeredBackends := []string{}
	for _, b := range backends {
		// Create new backend instance
		state.addBackend(b)
		registeredBackends = append(registeredBackends, b.ID)
	}

	// Remove node instances, no longer present
	// in the configuration.
	for _, b := range state.backends {
		present := false
		for _, beID := range registeredBackends {
			if beID == b.ID {
				present = true
				break
			}
		}
		if !present {
			state.removeBackend(b)
		}
	}

}

// GetBackendByID retrieves a backend by it's ID
func (state *State) GetBackendByID(id string) *Backend {
	for _, b := range state.backends {
		if b.ID == id {
			return b
		}
	}
	return nil
}

// AddBackend adds and starts a backend in the cluster. Thread safe.
func (state *State) AddBackend(backend *Backend) {
	state.mtx.Lock()
	defer state.mtx.Unlock()
	state.addBackend(backend)
}

// Unsafe addBackend to cluster
func (state *State) addBackend(backend *Backend) {
	current := state.GetBackendByID(backend.ID)
	if current == nil {
		// Just add the backend and start the instance
		state.backends = append(state.backends, backend)
		go backend.Start()
		return
	}

	// Replace instance when config changed
	if *(current.cfg) != *(backend.cfg) {
		backends := make([]*Backend, 0, len(state.backends))
		for _, b := range state.backends {
			if b == current {
				log.Println("Restarting backend:", b.ID)
				b.Stop()
				backends = append(backends, backend)
				go backend.Start()
			} else {
				backends = append(backends, b)
			}
		}
		state.backends = backends
	}
	// Noting else to do here
}

// RemoveBackend removes a backend from the cluster. Thread safe.
func (state *State) RemoveBackend(backend *Backend) {
	state.mtx.Lock()
	defer state.mtx.Unlock()
	state.removeBackend(backend)
}

// Unsafe remove backends from the controller.
func (state *State) removeBackend(backend *Backend) {
	backends := make([]*Backend, 0, len(state.backends))
	for _, b := range state.backends {
		if b == backend {
			b.Stop()
			continue
		}
		backends = append(backends, b)
	}
	state.backends = backends
}

// UpdateFrontends updates all frontends
func (state *State) UpdateFrontends(frontends []*Frontend) {
	state.mtx.Lock()
	defer state.mtx.Unlock()
	state.updateFrontends(frontends)
}

func (state *State) updateFrontends(frontends []*Frontend) {
	// Register all frontends from the config
	registeredIDs := []string{}
	for _, f := range frontends {
		// Registering is idempotent
		state.addFrontend(f)
		registeredIDs = append(registeredIDs, f.ID)
	}

	// Remove all frontends not longer in the config
	for _, frontend := range state.frontends {
		present := false
		for _, id := range registeredIDs {
			if frontend.ID == id {
				present = true
				break
			}
		}
		if !present {
			state.removeFrontend(frontend)
		}
	}
}

// AddFrontend adds a frontend to the cluster.
// This is an idempotent operation. If the frontend id
// is already registered, it will be replaced with
// the new frontend.
func (state *State) AddFrontend(frontend *Frontend) {
	state.mtx.Lock()
	defer state.mtx.Unlock()
	state.addFrontend(frontend)
}

// Unsafe interal add frontend
func (state *State) addFrontend(frontend *Frontend) {
	state.removeFrontend(frontend)
	state.frontends = append(state.frontends, frontend)
	log.Println("Registered frontend:", frontend.config.Key)
}

// RemoveFrontend removes a frontend from the cluster
func (state *State) RemoveFrontend(frontend *Frontend) {
	state.mtx.Lock()
	defer state.mtx.Unlock()
	state.removeFrontend(frontend)
}

// Unsafe internal removeFrontend without locking
func (state *State) removeFrontend(frontend *Frontend) {
	frontends := make([]*Frontend, 0, len(state.frontends))
	for _, f := range state.frontends {
		if f.ID == frontend.ID {
			continue
		}
		frontends = append(frontends, f)
	}
	state.frontends = frontends
}

// GetFrontendByID retrievs a frontend identified by
// its key from our list of frontends.
func (state *State) GetFrontendByID(id string) *Frontend {
	for _, f := range state.frontends {
		if f.ID == id {
			return f
		}
	}
	return nil
}

// GetFrontends retrievs all frontends in the controller
func (state *State) GetFrontends() []*Frontend {
	return state.frontends
}

// Log collects cluster information and writes
// them to the log.
func (state *State) Log() {
	log.Println("[Cluster state]")
}
