package v1

import (
	"net/http"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// The backend is either identified by ID or by
// hostname. The hostname must be an exact match.
func backendFromRequest(
	c echo.Context,
	tx pgx.Tx,
) (*store.BackendState, error) {
	ctx := c.(*APIContext)
	id := strings.TrimSpace(c.QueryParam("backend_id"))
	host := strings.TrimSpace(c.QueryParam("backend_host"))

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

	backend, err := store.GetBackendState(ctx.Ctx(), tx, q)
	if err != nil {
		return nil, err
	}

	return backend, nil
}

// BackendMeetingsList will retrieve all meetings for a
// given backend_id.
func BackendMeetingsList(c echo.Context) error {
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

	// Begin Query
	q := store.Q().Where("backend_id = ?", backend.ID)
	meetings, err := store.GetMeetingStates(cctx, tx, q)
	return c.JSON(http.StatusOK, meetings)
}

// BackendMeetingsEnd will stop all meetings for a
// given backend_id.
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
	res := map[string]interface{}{
		"cmd":   "end_all_meetings_request",
		"state": "queued",
	}
	return c.JSON(http.StatusAccepted, res)
}
