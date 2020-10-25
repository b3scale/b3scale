package cluster

// Command Creators

import (
	"errors"
	"time"

	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// Commands that can be handled by the controller
const (
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
		Deadline: time.Now().UTC().Add(5 * time.Minute),
	}
}
