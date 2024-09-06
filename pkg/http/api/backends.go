package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/store"
)

// ResourceBackends is a restful group for backend endpoints
var ResourceBackends = &Resource{
	List: RequireScope(
		auth.ScopeAdmin,
	)(apiBackendsList),

	Create: RequireScope(
		auth.ScopeAdmin,
		auth.ScopeNode,
	)(apiBackendCreate),

	Show: RequireScope(
		auth.ScopeAdmin,
		auth.ScopeNode,
	)(apiBackendShow),

	Update: RequireScope(
		auth.ScopeAdmin,
		auth.ScopeNode,
	)(apiBackendUpdate),

	Destroy: RequireScope(
		auth.ScopeAdmin,
	)(apiBackendDestroy),
}

// apiBackendsList will list all frontends known
// to the cluster or within the user scope.
// Admin scope is mandatory
func apiBackendsList(
	ctx context.Context,
	api *API,
) error {
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Begin Query
	q := store.Q()

	// Filter by host
	queryHost := api.QueryParam("host")
	if queryHost != "" {
		q = q.Where("host = ?", queryHost)
	}
	queryHostLike := api.QueryParam("host__like")
	if queryHostLike != "" {
		q = q.Where("host LIKE ?", fmt.Sprintf("%%%s%%", queryHostLike))
	}
	queryHostILike := api.QueryParam("host__ilike")
	if queryHostILike != "" {
		q = q.Where("host ILIKE ?", fmt.Sprintf("%%%s%%", queryHostILike))
	}

	// Set ordering
	q = q.OrderBy("backends.host ASC")

	backends, err := store.GetBackendStates(ctx, tx, q)
	return api.JSON(http.StatusOK, backends)
}

// BackendCreate will add a new backend to the cluster.
// Requires admin scope
func apiBackendCreate(
	ctx context.Context,
	api *API,
) error {
	b := &store.BackendState{}
	if err := api.Bind(b); err != nil {
		return err
	}

	// Only allow create with well known fields
	backend := store.InitBackendState(&store.BackendState{
		Backend:    b.Backend,
		Settings:   b.Settings,
		AdminState: b.AdminState,
		LoadFactor: b.LoadFactor,
	})

	if api.HasScope(auth.ScopeNode) {
		backend.AgentRef = &api.Ref
	}

	if err := backend.Validate(); err != nil {
		return err
	}

	// Begin transaction and save new backend state
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := backend.Save(ctx, tx); err != nil {
		return err
	}

	// Enqueue node refresh command
	cmd := cluster.UpdateNodeState(&cluster.UpdateNodeStateRequest{
		ID: backend.ID,
	})
	if err := store.QueueCommand(ctx, tx, cmd); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return api.JSON(http.StatusOK, backend)
}

// apiBackendShow will retrieve a single backend by ID.
// Requires admin scope
func apiBackendShow(
	ctx context.Context,
	api *API,
) error {
	id := api.Param("id")

	// Begin TX
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Begin Query
	q := store.Q().Where("id = ?", id)

	if api.HasScope(auth.ScopeNode) {
		q = q.Where("agent_ref = ?", api.Ref)
	}

	backend, err := store.GetBackendState(ctx, tx, q)

	if backend == nil {
		return echo.ErrNotFound
	}

	return api.JSON(http.StatusOK, backend)
}

// apiBackendDestroy will start a backend decommissioning.
// Requires admin scope.
func apiBackendDestroy(
	ctx context.Context,
	api *API,
) error {
	id := api.Param("id")
	force := config.IsEnabled(api.QueryParam("force"))

	// Begin TX
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Begin Query
	q := store.Q().Where("id = ?", id)
	if api.HasScope(auth.ScopeNode) {
		q = q.Where("agent_ref = ?", api.Ref)
	}

	backend, err := store.GetBackendState(ctx, tx, q)
	if backend == nil {
		return echo.ErrNotFound
	}

	if force {
		// force removal of backend. this is a hard delete
		// without decommissioning.
		if err := backend.Delete(ctx, tx); err != nil {
			return err
		}
		backend.AdminState = "destroyed"
	} else {
		// Request backend decommissioning.
		backend.AdminState = "decommissioned"
		if err := backend.Save(ctx, tx); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return api.JSON(http.StatusOK, backend)
}

// apiBackendUpdate will update the frontend with values
// provided by the request. Only keys provided will
// be updated.
func apiBackendUpdate(
	ctx context.Context,
	api *API,
) error {
	id := api.Param("id")
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Begin Query
	q := store.Q().Where("id = ?", id)
	if api.HasScope(auth.ScopeNode) {
		q = q.Where("agent_ref = ?", api.Ref)
	}

	update, err := store.GetBackendState(ctx, tx, q)
	if err != nil {
		return err
	}
	if update == nil {
		return echo.ErrNotFound
	}
	backend, err := store.GetBackendState(ctx, tx, q)
	if err != nil {
		return err
	}

	// Update backend
	if err := api.Bind(update); err != nil {
		return err
	}

	// Apply update for well known fields
	backend.Backend = update.Backend
	backend.Settings = update.Settings
	backend.AdminState = update.AdminState
	backend.LoadFactor = update.LoadFactor

	if err := backend.Validate(); err != nil {
		return err
	}

	// Persist updated backend
	if err := backend.Save(ctx, tx); err != nil {
		return err
	}

	// Enqueue node refresh command
	cmd := cluster.UpdateNodeState(&cluster.UpdateNodeStateRequest{
		ID: backend.ID,
	})
	if err := store.QueueCommand(ctx, tx, cmd); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return api.JSON(http.StatusOK, backend)
}
