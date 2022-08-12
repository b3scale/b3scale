package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/jackc/pgx/v4/pgxpool"
)

// AgentRPC api - because sometimes having things
// in a single transaction can be a good thing

// Errors
var (
	ErrInvalidAction = errors.New("the requsted RPC action is unknown")
)

// RPC status
const (
	RPCStatusOK    = "ok"
	RPCStatusError = "error"
)

// RPCPayload is a json raw message which then can be decoded
// into the actual request
type RPCPayload json.RawMessage

// RPCResult is anything the handle responds with
type RPCResult interface{}

// RPCRequest is an incomming request with an
// action and a payload.
type RPCRequest struct {
	Action  string     `json:"action"`
	Payload RPCPayload `json:"payload"`
}

// RPCHandler contains a database connection
type RPCHandler struct {
	Conn *pgxpool.Conn
}

// RPCResponse is the result of an RPC request
type RPCResponse struct {
	Status string    `json:"status"`
	Result RPCResult `json:"result"`
}

// Responses

// RPCError is an RPC error response
func RPCError(err error) *RPCResponse {
	res := &RPCResponse{
		Status: RPCStatusError,
		Result: err.Error(),
	}
	return res
}

// RPCSuccess is a successful RPC response
func RPCSuccess(result RPCResult) *RPCResponse {
	res := &RPCResponse{
		Status: RPCStatusOK,
		Result: result,
	}
	return res
}

// Actions
const (
	RPCMeetingStateReset     = "meeting_state_reset"
	RPCMeetingSetRunning     = "meeting_set_running"
	RPCMeetingAddAttendee    = "meeting_add_attendee"
	RPCMeetingRemoveAttendee = "meeting_remove_attendee"
)

// Payloads

// MeetingStateResetRequest contains a meetingID
type MeetingStateResetRequest struct {
	InternalMeetingID string `json:"internal_meeting_id"`
}

// MeetingSetRunningRequest contains a meetingID
type MeetingSetRunningRequest struct {
	InternalMeetingID string `json:"internal_meeting_id"`
	Running           bool   `json:"running"`
}

// MeetingAddAttendeeRequest will add an attedee to the
// attendees list of a meeting
type MeetingAddAttendeeRequest struct {
	InternalMeetingID string        `json:"internal_meeting_id"`
	Attendee          *bbb.Attendee `json:"attendee"`
}

// MeetingRemoveAttendeeRequest will remove an attendee from a meeting
// identified by the internal user id
type MeetingRemoveAttendeeRequest struct {
	InternalMeetingID string `json:"internal_meeting_id"`
	InternalUserID    string `json:"internal_user_id"`
}

// Dispatch will invoke the RPC handlers with the decoded
// request payload.
func (rpc *RPCRequest) Dispatch(
	ctx context.Context,
	handler *RPCHandler,
) *RPCResponse {
	var result RPCResult
	var err error

	switch rpc.Action {
	case RPCMeetingStateReset:
		req := &MeetingStateResetRequest{}
		if err := json.Unmarshal(rpc.Payload, &req); err != nil {
			return RPCError(err)
		}
		result, err = handler.MeetingStateReset(ctx, req)

	case RPCMeetingSetRunning:
		req := &MeetingSetRunningRequest{}
		if err := json.Unmarshal(rpc.Payload, &req); err != nil {
			return RPCError(err)
		}
		result, err = handler.MeetingSetRunning(ctx, req)

	case RPCMeetingAddAttendee:
		req := &MeetingAddAttendeeRequest{}
		if err := json.Unmarshal(rpc.Payload, &req); err != nil {
			return RPCError(err)
		}
		result, err = handler.MeetingAddAttendee(ctx, req)

	case RPCMeetingRemoveAttendee:
		req := &MeetingRemoveAttendeeRequest{}
		if err := json.Unmarshal(rpc.Payload, &req); err != nil {
			return RPCError(err)
		}
		result, err = handler.MeetingRemoveAttendee(ctx, req)

	default:
		err = ErrInvalidAction
	}
	if err != nil {
		return RPCError(err)
	}
	return RPCSuccess(result)
}

// Handler

// MeetingStateReset clears the attendees list and
// sets the running flag to false
func (rpc *RPCHandler) MeetingStateReset(
	ctx context.Context,
	req *MeetingStateResetRequest,
) (RPCResult, error) {
	return nil, nil
}

// MeetingSetRunning sets the meeting is running
// flag for a meeting. The meeting will be awaited.
func (rpc *RPCHandler) MeetingSetRunning(
	ctx context.Context,
	req *MeetingSetRunningRequest,
) (RPCResult, error) {
	return nil, nil
}

// MeetingAddAttendee insers an attendee into the list
func (rpc *RPCHandler) MeetingAddAttendee(
	ctx context.Context,
	req *MeetingAddAttendeeRequest,
) (RPCResult, error) {
	return nil, nil
}

// MeetingRemoveAttendee removes an attendee from the list
func (rpc *RPCHandler) MeetingRemoveAttendee(
	ctx context.Context,
	req *MeetingRemoveAttendeeRequest,
) (RPCResult, error) {
	return nil, nil
}

// HTTP API

// ResourceAgentRPC is the API resource for creating RPC requests
var ResourceAgentRPC = &Resource{
	// Create dispatches an RPC request
	Create: RequireScope(
		ScopeNode,
	)(func(ctx context.Context, api *API) error {
		// Decode request
		rpc := &RPCRequest{}
		if err := api.Bind(rpc); err != nil {
			return api.JSON(http.StatusBadRequest, RPCError(err))
		}

		// Execute op
		res := rpc.Dispatch(ctx, &RPCHandler{
			Conn: api.Conn,
		})

		// Make JSON response
		code := http.StatusOK
		if res.Status == RPCStatusError {
			code = http.StatusBadRequest
		}
		return api.JSON(code, res)
	}),
}
