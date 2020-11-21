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
	// Backend
	CmdRemoveBackend   = "remove_backend"
	CmdUpdateNodeState = "update_node_state"

	// Meetings
	CmdUpdateMeetingState = "update_meeting_state"
)

var (
	// ErrUnknownCommand indicates, that the command was not
	// understood by the controller.
	ErrUnknownCommand = errors.New("command unknown")
)

// AddFrontendRequest holds all params for creating a frontend
type AddFrontendRequest struct {
	Frontend *bbb.Frontend `json:"frontend"`
	Active   bool
}

// AddFrontend creates a new frontend in the state
func AddFrontend(req *AddFrontendRequest) *store.Command {
	return &store.Command{
		Action:   CmdAddFrontend,
		Params:   req,
		Deadline: store.NextDeadline(10 * time.Minute),
	}

}

// RemoveFrontendRequest holds an identifier
type RemoveFrontendRequest struct {
	ID string `json:"id"`
}

// RemoveFrontend creates a new command
func RemoveFrontend(req *RemoveFrontendRequest) *store.Command {
	return &store.Command{
		Action:   CmdRemoveFrontend,
		Params:   req,
		Deadline: store.NextDeadline(10 * time.Minute),
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

// UpdateNodeStateRequest requests a status update
// from a backend identified by ID
type UpdateNodeStateRequest struct {
	ID string // the backend state id
}

// UpdateNodeState creates a update status command
func UpdateNodeState(req *UpdateNodeStateRequest) *store.Command {
	return &store.Command{
		Action:   CmdUpdateNodeState,
		Params:   req,
		Deadline: store.NextDeadline(10 * time.Minute),
	}
}

// UpdateMeetingStateRequest requests the refresh of a meeting
type UpdateMeetingStateRequest struct {
	ID string // the meeting ID
}

// UpdateMeetingState makes a new meeting refresh command
func UpdateMeetingState(
	req *UpdateMeetingStateRequest,
) *store.Command {
	return &store.Command{
		Action:   CmdUpdateMeetingState,
		Params:   req,
		Deadline: store.NextDeadline(10 * time.Minute),
	}
}
