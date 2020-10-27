package cluster

// Command Creators

import (
	"errors"
	"time"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// Commands that can be handled by the controller
const (
	CmdAddBackend       = "add_backend"
	CmdRemoveBackend    = "remove_backend"
	CmdLoadBackendState = "load_backend_state"
)

var (
	// ErrUnknownCommand indicates, that the command was not
	// understood by the controller.
	ErrUnknownCommand = errors.New("command unknown")
)

// FetchBackendState initiates the loading of the
// entire state (meetings, recordings, text-tracks) from a backend
func FetchBackendState(backend *Backend) *store.Command {
	return &store.Command{
		Action:   CmdLoadBackendState,
		Params:   backend.state.ID,
		Deadline: store.NextDeadline(5 * time.Minute),
	}
}

// AddBackendRequest is a collection of params
// for creating a new backend state.
type AddBackendRequest struct {
	Backend *bbb.Backend `json:"backend"`
	Tags    []string     `json:"tags"`
}

// AddBackend inserts a new backend into
// the cluster state.
func AddBackend(req *AddBackendRequest) *store.Command {
	return &store.Command{
		Action:   CmdAddBackend,
		Params:   req,
		Deadline: store.NextDeadline(2 * time.Minute),
	}
}

// RemoveBackendRequest declares the removal
// of a backend node from the cluster state.
type RemoveBackendRequest struct {
	ID string `json:"id"`
}

// RemoveBackend will remove a given cluster
// backend from the state.
func RemoveBackend(req *RemoveBackendRequest) *store.Command {
	return &store.Command{
		Action:   CmdRemoveBackend,
		Params:   req,
		Deadline: store.NextDeadline(10 * time.Minute),
	}
}

// LoadBackendStateRequest describes intent loading
// an entire state from a bbb instance.
type LoadBackendStateRequest struct {
	ID string // the backend state id
}

// LoadBackendState creates a load state command
func LoadBackendState(req *LoadBackendStateRequest) *store.Command {
	return &store.Command{
		Action:   CmdLoadBackendState,
		Params:   req,
		Deadline: store.NextDeadline(10 * time.Minute),
	}
}
