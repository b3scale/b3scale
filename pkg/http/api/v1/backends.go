package v1

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// BackendsList will list all frontends known
// to the cluster or within the user scope.
// ! requires: `admin`
func BackendsList(c echo.Context) error {
	ctx := c.(*APIContext)
	reqCtx := ctx.Ctx()

	// Begin TX
	tx, err := store.ConnectionFromContext(reqCtx).Begin(reqCtx)
	if err != nil {
		return err
	}
	defer tx.Rollback(reqCtx)

	// Begin Query
	q := store.Q()

	// Filter by host
	queryHost := c.QueryParam("host")
	if queryHost != "" {
		q = q.Where("host = ?", queryHost)
	}
	queryHostLike := c.QueryParam("host__like")
	if queryHostLike != "" {
		q = q.Where("host LIKE ?", fmt.Sprintf("%%%s%%", queryHostLike))
	}

	backends, err := store.GetBackendStates(reqCtx, tx, q)
	return c.JSON(http.StatusOK, backends)
}

// BackendCreate will add a new backend to the cluster.
// ! requires: `admin`
func BackendCreate(c echo.Context) error {
	ctx := c.(*APIContext)
	reqCtx := ctx.Ctx()

	b := &store.BackendState{}
	if err := c.Bind(b); err != nil {
		return err
	}

	// Only allow create with well known fields
	backend := store.InitBackendState(&store.BackendState{
		Backend:    b.Backend,
		Settings:   b.Settings,
		AdminState: b.AdminState,
		LoadFactor: b.LoadFactor,
	})

	if err := backend.Validate(); err != nil {
		return err
	}

	// Begin transaction and save new backend state
	tx, err := store.ConnectionFromContext(reqCtx).Begin(reqCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
		return err
	}
	defer tx.Rollback(reqCtx)

	if err := backend.Save(reqCtx, tx); err != nil {
		return err
	}

	if err := tx.Commit(reqCtx); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, backend)
}

// BackendRetrieve will retrieve a single backend by ID.
// ! requires: `admin`
func BackendRetrieve(c echo.Context) error {
	ctx := c.(*APIContext)
	reqCtx := ctx.Ctx()

	id := c.Param("id")

	// Begin TX
	tx, err := store.ConnectionFromContext(reqCtx).Begin(reqCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
		return err
	}
	defer tx.Rollback(reqCtx)

	// Begin Query
	q := store.Q().Where("id = ?", id)
	backend, err := store.GetBackendState(reqCtx, tx, q)

	if backend == nil {
		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, backend)
}

// BackendDestroy will start a backend decommissioning.
// ! requires: `admin`
func BackendDestroy(c echo.Context) error {
	ctx := c.(*APIContext)
	reqCtx := ctx.Ctx()

	id := c.Param("id")

	force := config.IsEnabled(c.QueryParam("force"))

	// Begin TX
	tx, err := store.ConnectionFromContext(reqCtx).Begin(reqCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
		return err
	}
	defer tx.Rollback(reqCtx)

	// Begin Query
	q := store.Q().Where("id = ?", id)
	backend, err := store.GetBackendState(reqCtx, tx, q)
	if backend == nil {
		return echo.ErrNotFound
	}

	if force {
		// force removal of backend. this is a hard delete
		// without decommissioning.
		if err := backend.Delete(reqCtx, tx); err != nil {
			return err
		}
		backend.AdminState = "destroyed"

	} else {
		// Request backend decommissioning.
		backend.AdminState = "decommissioned"
		if err := backend.Save(reqCtx, tx); err != nil {
			return err
		}
	}

	if err := tx.Commit(reqCtx); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, backend)
}

// BackendUpdate will update the frontend with values
// provided by the request. Only keys provided will
// be updated.
// ! requires: `admin`
func BackendUpdate(c echo.Context) error {
	ctx := c.(*APIContext)
	reqCtx := ctx.Ctx()

	id := c.Param("id")

	// Begin TX
	tx, err := store.ConnectionFromContext(reqCtx).Begin(reqCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
		return err
	}
	defer tx.Rollback(reqCtx)

	// Begin Query
	q := store.Q().Where("id = ?", id)
	update, err := store.GetBackendState(reqCtx, tx, q)
	if err != nil {
		return err
	}
	if update == nil {
		return echo.ErrNotFound
	}

	backend, err := store.GetBackendState(reqCtx, tx, q)
	if err != nil {
		return err
	}

	// Update backend
	if err := c.Bind(update); err != nil {
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
	if err := backend.Save(reqCtx, tx); err != nil {
		return err
	}

	if err := tx.Commit(reqCtx); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, backend)
}
