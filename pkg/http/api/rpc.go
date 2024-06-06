package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/store"
)

// AgentRPC api - because sometimes having things
// in a single transaction can be a good thing

// Errors
var (
	ErrInvalidAction  = errors.New("the requsted RPC action is unknown")
	ErrInvalidBackend = errors.New("backend not associated with meeting")
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

// NewRPCRequest creates a new RPC request
func NewRPCRequest(
	action string,
	params interface{},
) *RPCRequest {
	payload, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}
	req := &RPCRequest{
		Action:  action,
		Payload: payload,
	}
	return req
}

// RPCHandler contains a database connection
type RPCHandler struct {
	AgentRef string
	Backend  *store.BackendState
	Conn     *pgxpool.Conn
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
	ActionMeetingStateReset     = "meeting_state_reset"
	ActionMeetingSetRunning     = "meeting_set_running"
	ActionMeetingAddAttendee    = "meeting_add_attendee"
	ActionMeetingRemoveAttendee = "meeting_remove_attendee"
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

// Action Creators

// RPCMeetingStateReset creates an meeting state reset request
func RPCMeetingStateReset(params *MeetingStateResetRequest) *RPCRequest {
	return NewRPCRequest(ActionMeetingStateReset, params)
}

// RPCMeetingSetRunning creates a set running request
func RPCMeetingSetRunning(params *MeetingSetRunningRequest) *RPCRequest {
	return NewRPCRequest(ActionMeetingSetRunning, params)
}

// RPCMeetingAddAttendee creates a new add attendee request
func RPCMeetingAddAttendee(params *MeetingAddAttendeeRequest) *RPCRequest {
	return NewRPCRequest(ActionMeetingAddAttendee, params)
}

// RPCMeetingRemoveAttendee creates a remove attendee request
func RPCMeetingRemoveAttendee(params *MeetingRemoveAttendeeRequest) *RPCRequest {
	return NewRPCRequest(ActionMeetingRemoveAttendee, params)
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
	case ActionMeetingStateReset:
		req := &MeetingStateResetRequest{}
		if err := json.Unmarshal(rpc.Payload, &req); err != nil {
			return RPCError(err)
		}
		result, err = handler.MeetingStateReset(ctx, req)

	case ActionMeetingSetRunning:
		req := &MeetingSetRunningRequest{}
		if err := json.Unmarshal(rpc.Payload, &req); err != nil {
			return RPCError(err)
		}
		result, err = handler.MeetingSetRunning(ctx, req)

	case ActionMeetingAddAttendee:
		req := &MeetingAddAttendeeRequest{}
		if err := json.Unmarshal(rpc.Payload, &req); err != nil {
			return RPCError(err)
		}
		result, err = handler.MeetingAddAttendee(ctx, req)

	case ActionMeetingRemoveAttendee:
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

// logMeetingNotFound creates a log message when the meeting
// could not be found within deadline.
func (rpc *RPCHandler) logMeetingNotFound(internalID string) {
	log.Warn().
		Str("agent", rpc.AgentRef).
		Str("backend", rpc.Backend.Backend.Host).
		Str("backend_id", rpc.Backend.ID).
		Str("internal_meeting_id", internalID).
		Msg("meeting not found within deadline")
}

// MeetingStateReset clears the attendees list and
// sets the running flag to false
func (rpc *RPCHandler) MeetingStateReset(
	ctx context.Context,
	req *MeetingStateResetRequest,
) (RPCResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	meeting, tx, err := store.AwaitMeetingState(ctx, rpc.Conn, store.Q().
		Where("meetings.backend_id = ?", rpc.Backend.ID).
		Where("meetings.internal_id = ?", req.InternalMeetingID))
	if errors.Is(err, context.DeadlineExceeded) {
		rpc.logMeetingNotFound(req.InternalMeetingID)
	}
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Update state
	meeting.Meeting.Running = false
	meeting.Meeting.Attendees = []*bbb.Attendee{}

	if err := meeting.Save(ctx, tx); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

// MeetingSetRunning sets the meeting is running
// flag for a meeting. The meeting will be awaited.
func (rpc *RPCHandler) MeetingSetRunning(
	ctx context.Context,
	req *MeetingSetRunningRequest,
) (RPCResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	meeting, tx, err := store.AwaitMeetingState(ctx, rpc.Conn, store.Q().
		Where("meetings.backend_id = ?", rpc.Backend.ID).
		Where("meetings.internal_id = ?", req.InternalMeetingID))
	if errors.Is(err, context.DeadlineExceeded) {
		rpc.logMeetingNotFound(req.InternalMeetingID)
	}
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Update state
	meeting.Meeting.Running = req.Running

	// Commit changes
	if err := meeting.Save(ctx, tx); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

// MeetingAddAttendee insers an attendee into the list
func (rpc *RPCHandler) MeetingAddAttendee(
	ctx context.Context,
	req *MeetingAddAttendeeRequest,
) (RPCResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	meeting, tx, err := store.AwaitMeetingState(ctx, rpc.Conn, store.Q().
		Where("meetings.backend_id = ?", rpc.Backend.ID).
		Where("meetings.internal_id = ?", req.InternalMeetingID))
	if errors.Is(err, context.DeadlineExceeded) {
		rpc.logMeetingNotFound(req.InternalMeetingID)
	}
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Update state
	attendees := meeting.Meeting.Attendees
	if attendees == nil {
		attendees = []*bbb.Attendee{}
	}
	attendees = append(attendees, req.Attendee)
	meeting.Meeting.Attendees = attendees

	if err := meeting.Save(ctx, tx); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

// MeetingRemoveAttendee removes an attendee from the list
func (rpc *RPCHandler) MeetingRemoveAttendee(
	ctx context.Context,
	req *MeetingRemoveAttendeeRequest,
) (RPCResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	meeting, tx, err := store.AwaitMeetingState(ctx, rpc.Conn, store.Q().
		Where("meetings.backend_id = ?", rpc.Backend.ID).
		Where("meetings.internal_id = ?", req.InternalMeetingID))
	if errors.Is(err, context.DeadlineExceeded) {
		rpc.logMeetingNotFound(req.InternalMeetingID)
	}
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Update state
	attendees := meeting.Meeting.Attendees
	if attendees == nil {
		return nil, nil // nothing to do here...
	}
	filtered := make([]*bbb.Attendee, 0, len(meeting.Meeting.Attendees))
	for _, a := range meeting.Meeting.Attendees {
		if a.InternalUserID == req.InternalUserID {
			continue // The user just left
		}
		filtered = append(filtered, a)
	}
	meeting.Meeting.Attendees = filtered

	if err := meeting.Save(ctx, tx); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return nil, nil
}

// HTTP API

// ResourceAgentRPC is the API resource for creating RPC requests
var ResourceAgentRPC = &Resource{
	// Create dispatches an RPC request
	Create: RequireScope(
		auth.ScopeNode,
	)(func(ctx context.Context, api *API) error {
		// Decode request
		rpc := &RPCRequest{}
		if err := api.Bind(rpc); err != nil {
			return api.JSON(http.StatusBadRequest, RPCError(err))
		}

		tx, err := api.Conn.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		// Get current backend
		backend, err := BackendFromAgentRef(ctx, api, tx)
		if err != nil {
			return err
		}
		if backend == nil {
			return echo.ErrForbidden // We require an active agent
		}
		tx.Rollback(ctx) // Transaction is not longer required

		// Execute op
		res := rpc.Dispatch(ctx, &RPCHandler{
			AgentRef: api.Ref,
			Backend:  backend,
			Conn:     api.Conn,
		})

		// Make JSON response. We do not use HTTP status
		// here for error signaling, as this will be decoded
		// as an APIError.
		return api.JSON(http.StatusOK, res)
	}),
}
