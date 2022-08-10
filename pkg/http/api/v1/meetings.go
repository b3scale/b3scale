package v1

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/b3scale/b3scale/pkg/store"
)

// APIResourceMeetings is a restful group for meetings
var APIResourceMeetings = &APIResource{
	List: RequireScope(
		ScopeAdmin,
	)(apiMeetingsList),

	Show: RequireScope(
		ScopeAdmin,
		ScopeNode,
	)(apiMeetingsShow),
}

// apiMeetingsList will retrieve all meetings within the scope
// of a given backend identified by ID. Limiting to the backend
// scope is important because the returned list might be long.
func apiMeetingsList(
	ctx context.Context,
	api *APIContext,
) error {
	// Begin TX
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	backend, err := BackendFromQuery(ctx, api, tx)
	if err != nil {
		return err
	}

	if backend == nil {
		return echo.ErrNotFound
	}

	// Begin Query
	q := store.Q().Where("backend_id = ?", backend.ID)
	meetings, err := store.GetMeetingStates(ctx, tx, q)
	if err != nil {
		return err
	}

	return api.JSON(http.StatusOK, meetings)
}

// apiMeetingsShow will get a single meeting by ID
// The internal meeting ID can be used, by prefixing the
// ID parameter with an `internal:`.
//
// This inband signaling is a compromise.
func apiMeetingsShow(
	ctx context.Context,
	api *APIContext,
) error {
	// Begin TX
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	backend, err := BackendFromAgentRef(ctx, api, tx)
	if err != nil {
		return err
	}

	// The backend must be available if the scope is node
	if api.HasScope(ScopeNode) && backend == nil {
		return echo.ErrForbidden
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
		return err
	}

	return api.JSON(http.StatusOK, meeting)
}
