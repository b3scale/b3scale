package v1

import (
	"context"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"

	"github.com/b3scale/b3scale/pkg/store"
)

// The backend is either identified by ID or by
// hostname. The hostname must be an exact match.
func backendFromRequest(
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

// apiMeetingsList will retrieve all meetings within the scope
// of a given backend identified by ID.
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

	backend, err := backendFromRequest(ctx, api, tx)
	if err != nil {
		return err
	}

	if backend == nil {
		return echo.ErrNotFound
	}

	// Begin Query
	q := store.Q().Where("backend_id = ?", backend.ID)
	meetings, err := store.GetMeetingStates(ctx, tx, q)
	return api.JSON(http.StatusOK, meetings)
}

// BackendMeetingsEndResponse is the result of the end
// all meeting on backend request and will indicate
// that the command was queued and the request was
// accepted.

// BackendMeetingsEnd will stop all meetings for a
// given backend_id.
/*
func BackendMeetingsEnd(c echo.Context) error {
	ctx := c.(*APIContext)
	cctx := ctx.Ctx()

	// Begin TX
	tx, err := store.ConnectionFromContext(cctx).Begin(cctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(cctx)

	backend, err := backendFromRequest(c, tx)
	if err != nil {
		return err
	}

	if backend == nil {
		return echo.ErrNotFound
	}

	cmd := cluster.EndAllMeetings(&cluster.EndAllMeetingsRequest{
		BackendID: backend.ID,
	})
	if err := store.QueueCommand(cctx, tx, cmd); err != nil {
		return err
	}

	if err := tx.Commit(cctx); err != nil {
		return err
	}

	// Make response
	return c.JSON(http.StatusAccepted, cmd)
}
*/
