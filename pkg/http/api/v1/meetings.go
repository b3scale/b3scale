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

	backend, err := backendFromQuery(ctx, api, tx)
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
