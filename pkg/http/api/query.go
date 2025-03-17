package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/store"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
)

// BackendFromQuery resolves the backend, either identified by ID or by
// hostname. The hostname must be an exact match.
func BackendFromQuery(
	ctx context.Context,
	api *API,
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
	api *API,
	tx pgx.Tx,
) (*store.BackendState, error) {
	if !api.HasScope(auth.ScopeNode) {
		return nil, nil // does not apply
	}
	q := store.Q().Where("agent_ref = ?", api.Ref)
	return store.GetBackendState(ctx, tx, q)
}

// MeetingFromRequest resolves the current meeting
func MeetingFromRequest(
	ctx context.Context,
	api *API,
	tx pgx.Tx,
) (*store.MeetingState, error) {
	q := store.Q()

	backend, err := BackendFromAgentRef(ctx, api, tx)
	if err != nil {
		return nil, err
	}
	// The backend must be available if the scope is node
	if api.HasScope(auth.ScopeNode) && backend == nil {
		return nil, echo.ErrForbidden
	}

	id, internal := api.ParamID()
	if internal {
		q = q.Where("meetings.internal_id = ?", id)
	} else {
		q = q.Where("meetings.id = ?", id)
	}

	if api.HasScope(auth.ScopeNode) {
		q = q.Where("meetings.backend_id = ?", backend.ID)
	}

	meeting, err := store.GetMeetingState(ctx, tx, q)
	if err != nil {
		return nil, err
	}
	if meeting == nil {
		return nil, echo.ErrNotFound
	}
	return meeting, nil
}

// FrontendFromQueryParams retrieves the frontend from the store
// identified either by key or id. ID has precedence over key.
func FrontendFromQueryParams(
	ctx context.Context,
	api *API,
	tx pgx.Tx,
) (*store.FrontendState, error) {
	var (
		fe  *store.FrontendState
		err error
	)

	feID := api.QueryParam("frontend_id")
	feKey := api.QueryParam("frontend_key")

	if feID == "" && feKey == "" {
		return nil, fmt.Errorf("frontend_id and frontend_key are missing")
	}

	// When empty, get by key.
	if feID == "" {
		fe, err = store.GetFrontendStateByKey(ctx, tx, feKey)
		if err != nil {
			return nil, err
		}
	} else { // Otherwise use the ID
		fe, err = store.GetFrontendStateByID(ctx, tx, feID)
		if err != nil {
			return nil, err
		}
	}

	// Check if we could find a frontend
	if fe == nil {
		return nil, fmt.Errorf("a frontend could not be found")
	}

	return fe, nil
}
