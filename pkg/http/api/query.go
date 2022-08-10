package v1

import (
	"context"
	"strings"

	"github.com/b3scale/b3scale/pkg/store"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
)

// BackendFromQuery resolves the backend, either identified by ID or by
// hostname. The hostname must be an exact match.
func BackendFromQuery(
	ctx context.Context,
	api *APIContext,
	tx pgx.Tx,
) (*store.BackendState, error) {
	id := strings.TrimSpace(api.QueryParam("backend_id"))
	host := strings.TrimSpace(api.QueryParam("backend_host"))

	hasQuery := false
	q := store.Q()
	if id != "" {
		q = q.Where("id = ?", id)
		hasQuery = true
	}
	if host != "" {
		q = q.Where("host = ?", host)
		hasQuery = true
	}
	if !hasQuery {
		return nil, echo.ErrBadRequest
	}

	backend, err := store.GetBackendState(ctx, tx, q)
	if err != nil {
		return nil, err
	}
	return backend, nil
}

// BackendFromAgentRef resolves the backend attached to
// the current node agent.
func BackendFromAgentRef(
	ctx context.Context,
	api *APIContext,
	tx pgx.Tx,
) (*store.BackendState, error) {
	if !api.HasScope(ScopeNode) {
		return nil, nil // does not apply
	}
	q := store.Q().Where("agent_ref = ?", api.Ref)
	return store.GetBackendState(ctx, tx, q)
}

// MeetingFromRequest resolves the current meeting
func MeetingFromRequest(
	ctx context.Context,
	api *APIContext,
	tx pgx.Tx,
) (*store.MeetingState, error) {
	backend, err := BackendFromAgentRef(ctx, api, tx)
	if err != nil {
		return nil, err
	}

	// The backend must be available if the scope is node
	if api.HasScope(ScopeNode) && backend == nil {
		return nil, echo.ErrForbidden
	}

	id, internal := api.ParamID()
	q := store.Q()
	if internal {
		q = q.Where("meetings.internal_id = ?", id)
	} else {
		q = q.Where("meetings.id = ?", id)
	}

	if api.HasScope(ScopeNode) {
		q = q.Where("meetings.backend_id = ?", backend.ID)
	}

	meeting, err := store.GetMeetingState(ctx, tx, q)
	if err != nil {
		return nil, err
	}
	return meeting, nil
}
