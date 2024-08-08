package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/store"
)

// ResourceMeetings is a restful group for meetings
var ResourceMeetings = &Resource{
	List: RequireScope(
		auth.ScopeAdmin,
	)(apiMeetingsList),

	Show: RequireScope(
		auth.ScopeAdmin,
		auth.ScopeNode,
	)(apiMeetingShow),

	Update: RequireScope(
		auth.ScopeAdmin,
		auth.ScopeNode,
	)(apiMeetingUpdate),

	Destroy: RequireScope(
		auth.ScopeAdmin,
		auth.ScopeNode,
	)(apiMeetingDestroy),
}

// apiMeetingsList will retrieve all meetings within the scope
// of a given backend identified by ID. Limiting to the backend
// scope is important because the returned list might be long.
func apiMeetingsList(
	ctx context.Context,
	api *API,
) error {
	// Begin TX
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	id := strings.TrimSpace(api.QueryParam("backend_id"))
	host := strings.TrimSpace(api.QueryParam("backend_host"))

	// Begin Query
	q := store.Q()

	if id != "" || host != "" {
		backend, err := BackendFromQuery(ctx, tx, id, host)
		if err != nil {
			return err
		}
		if backend == nil {
			return echo.ErrNotFound
		}
		q = q.Where("backend_id = ?", backend.ID)
	}

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
func apiMeetingShow(
	ctx context.Context,
	api *API,
) error {
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	meeting, err := MeetingFromRequest(ctx, api, tx)
	if err != nil {
		return err
	}

	return api.JSON(http.StatusOK, meeting)
}

// apiMeetingsUpdate will update a meeting
func apiMeetingUpdate(
	ctx context.Context,
	api *API,
) error {
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	meeting, err := MeetingFromRequest(ctx, api, tx)
	if err != nil {
		return err
	}
	update, err := MeetingFromRequest(ctx, api, tx)
	if err != nil {
		return err
	}

	// Apply updates
	if err := api.Bind(update); err != nil {
		return err
	}

	// Only allow update of meeting data
	meeting.Meeting = update.Meeting

	if err := meeting.Save(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return api.JSON(http.StatusOK, meeting)
}

// apiMeetingDestroy will delete a meeting from the store
func apiMeetingDestroy(
	ctx context.Context,
	api *API,
) error {
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	meeting, err := MeetingFromRequest(ctx, api, tx)
	if err != nil {
		return err
	}

	if err := store.DeleteMeetingStateByID(ctx, tx, meeting.ID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return api.JSON(http.StatusOK, meeting)
}
