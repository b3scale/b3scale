package api

import (
	"context"
	"net/http"

	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/store"

	"github.com/labstack/echo/v4"
)

// ResourceAgentBackend is the resource for retrieving
// the currently with the agent associated backend
var ResourceAgentBackend = &Resource{
	List: RequireScope(
		auth.ScopeNode,
	)(apiAgentBackendShow),
}

// Get the active backend for the agent or 404
func apiAgentBackendShow(
	ctx context.Context,
	api *API,
) error {
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := store.Q().Where("agent_ref = ?", api.Ref)
	backend, err := store.GetBackendState(ctx, tx, q)
	if err != nil {
		return err
	}
	if backend == nil {
		return echo.ErrNotFound
	}
	return api.JSON(http.StatusOK, backend)
}

// ResourceAgentHeartbeat is the resource for receiving
// an agent heartbeat
var ResourceAgentHeartbeat = &Resource{
	Create: RequireScope(
		auth.ScopeNode,
	)(apiAgentHeartbeatCreate),
}

// Update the backends agent heartbeat
func apiAgentHeartbeatCreate(
	ctx context.Context,
	api *API,
) error {
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Begin Query
	q := store.Q().Where("agent_ref = ?", api.Ref)
	backend, err := store.GetBackendState(ctx, tx, q)
	if err != nil {
		return err
	}
	if backend == nil {
		return echo.ErrNotFound
	}

	heartbeat, err := backend.UpdateAgentHeartbeat(ctx, tx)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return api.JSON(http.StatusOK, heartbeat)
}
