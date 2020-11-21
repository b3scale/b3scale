package cluster

// Command Creators

import (
	"errors"
	"time"

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
