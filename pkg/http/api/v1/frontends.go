package v1

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/b3scale/b3scale/pkg/store"
)

//		ref := ctx.Context.QueryParam("subject_ref")
// FrontendsList will list all frontends known
// to the cluster or within the user scope.

func apiFrontendsList(
	ctx context.Context,
	api *APIContext,
) error {
	q := store.Q()
	// Force filters if not admin account
	if !api.HasScope(ScopeAdmin) {
		q = q.Where("account_ref = ?", api.Ref)
	}

	// Query parameter filters
	queryRef := api.QueryParam("subject_ref")
	if queryRef != "" {
		q = q.Where("account_ref = ?", queryRef)
	}
	queryKey := api.QueryParam("key")
	if queryKey != "" {
		q = q.Where("key = ?", queryKey)
	}
	queryKeyLike := api.QueryParam("key__like")
	if queryKeyLike != "" {
		q = q.Where("key LIKE ?", fmt.Sprintf("%%%s%%", queryKeyLike))
	}

	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	frontends, err := store.GetFrontendStates(ctx, tx, q)
	if err != nil {
		return err
	}
	return api.JSON(http.StatusOK, frontends)
}

// apiFrontendCreate will add a new frontend to the cluster.
// Admin scope is mandatory.
func apiFrontendCreate(
	ctx context.Context,
	api *APIContext,
) error {
	if !api.HasScope(ScopeAdmin) {
		return ErrScopeRequired(ScopeAdmin)
	}

	f := &store.FrontendState{}
	if err := api.Bind(f); err != nil {
		return err
	}

	frontend := store.InitFrontendState(&store.FrontendState{
		Frontend:   f.Frontend,
		Settings:   f.Settings,
		Active:     f.Active,
		AccountRef: f.AccountRef,
	})

	if err := frontend.Validate(); err != nil {
		return err
	}

	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := frontend.Save(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return api.JSON(http.StatusOK, frontend)
}

// apiFrontendShow will retrieve a single frontend
// identified by ID.
func apiFrontendShow(
	ctx context.Context,
	api *APIContext,
) error {
	id := api.Param("id")

	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := store.Q().Where("id = ?", id)
	if !api.HasScope(ScopeAdmin) {
		q = q.Where("account_ref = ?", api.Ref)
	}
	frontend, err := store.GetFrontendState(ctx, tx, q)
	if err != nil {
		return err
	}
	if frontend == nil {
		return echo.ErrNotFound
	}
	return api.JSON(http.StatusOK, frontend)
}

// apiFrontendDestroy will remove a frontend from the cluster.
// The frontend is identified by ID.
// Admin scope is mandatory.
func apiFrontendDestroy(
	ctx context.Context,
	api *APIContext,
) error {
	id := api.Param("id")

	if !api.HasScope(ScopeAdmin) {
		return ErrScopeRequired(ScopeAdmin)
	}

	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := store.Q().Where("id = ?", id)
	frontend, err := store.GetFrontendState(ctx, tx, q)
	if err != nil {
		return err
	}
	if frontend == nil {
		return echo.ErrNotFound
	}

	if err := frontend.Delete(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	frontend.Active = false
	return api.JSON(http.StatusOK, frontend)
}

// apiFrontendUpdate will update the frontend with values
// provided by the request. Only keys provided will
// be updated.
func apiFrontendUpdate(
	ctx context.Context,
	api *APIContext,
) error {
	id := api.Param("id")

	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := store.Q().Where("id = ?", id)
	if !api.HasScope(ScopeAdmin) {
		q = q.Where("account_ref = ?", api.Ref)
	}

	frontend, err := store.GetFrontendState(ctx, tx, q)
	if err != nil {
		return err
	}
	if frontend == nil {
		return echo.ErrNotFound
	}

	update, err := store.GetFrontendState(ctx, tx, q)
	if err != nil {
		return err
	}

	if err := api.Bind(update); err != nil {
		return err
	}

	// Update fields
	frontend.Frontend = update.Frontend
	frontend.Active = update.Active
	frontend.Settings = update.Settings

	if api.HasScope(ScopeAdmin) {
		frontend.AccountRef = update.AccountRef
	}

	if err := frontend.Validate(); err != nil {
		return err
	}
	if err := frontend.Save(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return api.JSON(http.StatusOK, frontend)
}
