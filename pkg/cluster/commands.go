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
	CmdUpdateNodeState     = "update_node_state"
	CmdDecommissionBackend = "decommission_backend"

	// Meetings
	CmdUpdateMeetingState = "update_meeting_state"
	CmdEndAllMeetings     = "end_all_meetings"

	// Maintenance
	CmdCollectGarbage = "collect_garbage"
)

var (
	// ErrUnknownCommand indicates, that the command was not
	// understood by the controller.
	ErrUnknownCommand = errors.New("command unknown")
)

// DecommissionBackendRequest declares the removal
// of a backend node from the cluster state.
type DecommissionBackendRequest struct {
	ID string `json:"id"`
}

// DecommissionBackend will remove a given cluster
// backend from the state.
func DecommissionBackend(req *DecommissionBackendRequest) *store.Command {
	return &store.Command{
		Action:   CmdDecommissionBackend,
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

// EndAllMeetingsRequest contains parameters for the end
// all meetings command.
type EndAllMeetingsRequest struct {
	BackendID string
}

// EndAllMeetings will send end meeting api requests to all running meetings
// on a backend. This can be usefull to force decommissioning.
func EndAllMeetings(req *EndAllMeetingsRequest) *store.Command {
	return &store.Command{
		Action:   CmdEndAllMeetings,
		Params:   req,
		Deadline: store.NextDeadline(5 * time.Minute),
	}
}

// CollectGarbage requests removing stale states.
func CollectGarbage() *store.Command {
	return &store.Command{
		Action:   CmdCollectGarbage,
		Deadline: store.NextDeadline(5 * time.Minute),
	}
}
