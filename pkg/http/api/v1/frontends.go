package v1

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/b3scale/b3scale/pkg/store"
)

// FrontendsList will list all frontends known
// to the cluster or within the user scope.
func FrontendsList(c echo.Context) error {
	ctx := c.(*APIContext)
	ref := ctx.FilterAccountRef() // This will limit the scope to the `sub`
	reqCtx := ctx.Ctx()

	q := store.Q()

	// Apply filters
	if ref != nil {
		// this should only be the case if admin
		q = q.Where("account_ref = ?", *ref)
	}
	queryKey := c.QueryParam("key")
	if queryKey != "" {
		q = q.Where("key = ?", queryKey)
	}
	queryKeyLike := c.QueryParam("key__like")
	if queryKeyLike != "" {
		q = q.Where("key LIKE ?", fmt.Sprintf("%%%s%%", queryKeyLike))
	}

	tx, err := store.ConnectionFromContext(reqCtx).Begin(reqCtx)
	if err != nil {
		return err
	}
	defer tx.Rollback(reqCtx)
	frontends, err := store.GetFrontendStates(reqCtx, tx, q)
	return c.JSON(http.StatusOK, frontends)
}

// FrontendCreate will add a new frontend to the cluster.
func FrontendCreate(c echo.Context) error {
	ctx := c.(*APIContext)
	cctx := ctx.Ctx()
	isAdmin := ctx.HasScope(ScopeAdmin)
	accountRef := ctx.AccountRef()

	f := &store.FrontendState{}
	if err := c.Bind(f); err != nil {
		return err
	}

	frontend := store.InitFrontendState(&store.FrontendState{
		Frontend: f.Frontend,
		Settings: f.Settings,
		Active:   f.Active,
	})

	if isAdmin {
		frontend.AccountRef = f.AccountRef
	} else {
		frontend.AccountRef = &accountRef
	}

	if err := frontend.Validate(); err != nil {
		return err
	}

	tx, err := store.ConnectionFromContext(cctx).Begin(cctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(cctx)

	if err := frontend.Save(cctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(cctx); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, frontend)
}

// FrontendRetrieve will retrieve a single frontend
// identified by ID.
func FrontendRetrieve(c echo.Context) error {
	ctx := c.(*APIContext)
	cctx := ctx.Ctx()
	isAdmin := ctx.HasScope(ScopeAdmin)
	accountRef := ctx.AccountRef()
	id := c.Param("id")

	tx, err := store.ConnectionFromContext(cctx).Begin(cctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(cctx)

	q := store.Q().Where("id = ?", id)
	if !isAdmin {
		q = q.Where("account_ref = ?", accountRef)
	}

	frontend, err := store.GetFrontendState(cctx, tx, q)
	if err != nil {
		return err
	}

	if frontend == nil {
		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, frontend)
}

// FrontendDestroy will remove a frontend from the cluster.
// The frontend is identified by ID.
func FrontendDestroy(c echo.Context) error {
	ctx := c.(*APIContext)
	cctx := ctx.Ctx()
	isAdmin := ctx.HasScope(ScopeAdmin)
	accountRef := ctx.AccountRef()
	id := c.Param("id")

	tx, err := store.ConnectionFromContext(cctx).Begin(cctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(cctx)

	q := store.Q().Where("id = ?", id)
	if !isAdmin {
		q = q.Where("account_ref = ?", accountRef)
	}

	frontend, err := store.GetFrontendState(cctx, tx, q)
	if err != nil {
		return err
	}

	if frontend == nil {
		return echo.ErrNotFound
	}

	if err := frontend.Delete(cctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(cctx); err != nil {
		return err
	}

	frontend.Active = false
	return c.JSON(http.StatusOK, frontend)
}

// FrontendUpdate will update the frontend with values
// provided by the request. Only keys provided will
// be updated.
func FrontendUpdate(c echo.Context) error {
	ctx := c.(*APIContext)
	cctx := ctx.Ctx()
	isAdmin := ctx.HasScope(ScopeAdmin)
	accountRef := ctx.AccountRef()
	id := c.Param("id")

	tx, err := store.ConnectionFromContext(cctx).Begin(cctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(cctx)

	q := store.Q().Where("id = ?", id)
	if !isAdmin {
		q = q.Where("account_ref = ?", accountRef)
	}

	frontend, err := store.GetFrontendState(cctx, tx, q)
	if err != nil {
		return err
	}
	if frontend == nil {
		return echo.ErrNotFound
	}

	update, err := store.GetFrontendState(cctx, tx, q)
	if err != nil {
		return err
	}

	if err := c.Bind(update); err != nil {
		return err
	}

	// Update fields
	frontend.Frontend = update.Frontend
	frontend.Active = update.Active
	frontend.Settings = update.Settings

	if isAdmin {
		frontend.AccountRef = update.AccountRef
	}

	if err := frontend.Validate(); err != nil {
		return err
	}
	if err := frontend.Save(cctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(cctx); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, frontend)
}
