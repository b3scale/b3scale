package api

import (
	"context"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/rs/zerolog/log"

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
	if !api.HasScope(ScopeNode) {
		return nil, nil // does not apply
	}
	q := store.Q().Where("agent_ref = ?", api.Ref)
	return store.GetBackendState(ctx, tx, q)
}

// MeetingQueryFromRequest will build the meeting
// filters query.
func MeetingQueryFromRequest(
	ctx context.Context,
	api *API,
	tx pgx.Tx,
) (sq.SelectBuilder, error) {
	q := store.Q()

	backend, err := BackendFromAgentRef(ctx, api, tx)
	if err != nil {
		return q, err
	}

	// The backend must be available if the scope is node
	if api.HasScope(ScopeNode) && backend == nil {
		return q, echo.ErrForbidden
	}

	id, internal := api.ParamID()
	if internal {
		q = q.Where("meetings.internal_id = ?", id)
	} else {
		q = q.Where("meetings.id = ?", id)
	}

	if api.HasScope(ScopeNode) {
		q = q.Where("meetings.backend_id = ?", backend.ID)
	}
	return q, nil
}

// MeetingFromRequest resolves the current meeting
func MeetingFromRequest(
	ctx context.Context,
	api *API,
	tx pgx.Tx,
) (*store.MeetingState, error) {
	q, err := MeetingQueryFromRequest(ctx, api, tx)
	if err != nil {
		return nil, err
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

// AwaitMeetingFromRequest polls the database until the
// context is expired or the meeting showed up.
//
// Sometime we have to be aware, that the meeting
// might not yet be in the scaler state - however,
// bbb already fired an event for this.
//
// As all events are processed in their own goroutine,
// we sit back, wait, and poll the database
// for the internal meeting to come up. We should give
// up after 2-5 seconds.
func AwaitMeetingFromRequest(
	ctx context.Context,
	api *API,
	q sq.SelectBuilder,
) (*store.MeetingState, error) {
	for {
		if err := ctx.Err(); err != nil {
			id, internal := api.ParamID()
			log.Warn().
				Str("meeting", id).
				Bool("internal_meeting_id", internal).
				Msg("await meeting failed - context expired")
			return nil, nil // Context invalid or timed out, this is "OK"
		}

		tx, err := api.Conn.Begin(ctx)
		if err != nil {
			time.Sleep(150 * time.Millisecond)
			continue
		}
		defer tx.Rollback(ctx)

		// Try to get the meeting
		meeting, err := store.GetMeetingState(ctx, tx, q)
		if err != nil {
			return nil, err
		}
		if meeting != nil { // Found meeting!
			return meeting, nil
		}

		tx.Rollback(ctx) // Close transaction

		// Okay we should wait and try again..
		time.Sleep(150 * time.Millisecond)
	}
}
