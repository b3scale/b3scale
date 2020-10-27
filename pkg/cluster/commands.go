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

// AddBackend inserts a new backend into
// the cluster state.
func AddBackend(b *bbb.Backend) *store.Command {
	return &store.Command{
		Action:   CmdAddBackend,
		Params:   b,
		Deadline: store.NextDeadline(10 * time.Minute),
	}
}

// RemoveBackendByID will remove a given cluster
// backend from the state.
func RemoveBackendByID(id string) *store.Command {
	return &store.Command{
		Action:   CmdRemoveBackendByID,
		Params:   id,
		Deadline: store.NextDeadline(10 * time.Minute),
	}
}
