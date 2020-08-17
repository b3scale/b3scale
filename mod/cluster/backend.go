package cluster

import (
	"gitlab.com/infra.run/public/b3scale/mod/config"
)

const (
	// BackendStateInit when syncing the state
	BackendStateInit = iota
	// BackendStateReady when we accept requests
	BackendStateReady
	// BackendStateError when we do not accept requests
	BackendStateError
)

// The BackendState can be init, active or error
type BackendState int

// A Backend is a BigBlueButton instance and a node in
// the cluster.
//
// It has a host and a secret for request authentication.
// It syncs it's state with the bbb instance.
type Backend struct {
	State     BackendState
	LastError string

	config *config.Backend
}

// NewBackend creates a cluster node.
func NewBackend(config *config.Backend) *Backend {
	return &Backend{config: config}
}
