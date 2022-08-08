package v1

import (
	"context"
	"net/http"
)

// APIResourceCommands bundles read and create operations
// for manipulating the command queue.
var APIResourceCommands = &APIResource{
	List: RequireScope(
		ScopeAdmin,
	)(apiCommandsList),

	Create: RequireScope(
		ScopeAdmin,
	)(apiCommandCreate),
}

// apiCommandsList returns the current command queue
func apiCommandsList(ctx context.Context, api *APIContext) error {
	return api.JSON(http.StatusOK, []string{})
}

// apiCommandCreate adds a new well known command to the queue
func apiCommandCreate(ctx context.Context, api *APIContext) error {
	return api.JSON(http.StatusOK, "COMMAND")
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
