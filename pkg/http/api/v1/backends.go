package v1

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
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
		log.Fatal().Err(err).Msg("could not start transaction")
		return err
	}
	defer tx.Rollback(reqCtx)

	// Begin Query
	q := store.Q()
	backends, err := store.GetBackendStates(reqCtx, tx, q)
	c.JSON(http.StatusOK, backends)

	return nil
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

	// Force defaults
	b.ID = ""
	b.NodeState = ""
	b = store.InitBackendState(b)
	log.Print("B:", b)

	// Begin transaction and save new backend state
	tx, err := store.ConnectionFromContext(reqCtx).Begin(reqCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
		return err
	}
	defer tx.Rollback(reqCtx)

	if err := b.Save(reqCtx, tx); err != nil {
		return err
	}

	if err := tx.Commit(reqCtx); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, b)
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

	c.JSON(http.StatusOK, backend)

	return nil
}

// BackendDestroy will start a backend decommissioning.
// ! requires: `admin`
func BackendDestroy(c echo.Context) error {
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

	// Request backend decommissioning
	backend.AdminState = "decommissioned"
	if err := backend.Save(reqCtx, tx); err != nil {
		return err
	}

	if err := tx.Commit(reqCtx); err != nil {
		return err
	}

	c.JSON(http.StatusOK, backend)

	return nil
}

// BackendUpdate will update the frontend with values
// provided by the request. Only keys provided will
// be updated.
// ! requires: `admin`
func BackendUpdate(c echo.Context) error {
	ctx := c.(*APIContext)
	reqCtx := ctx.Ctx()

	id := c.Param("id")
	log.Info().Str("id", id).Msg("update backend")

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

	// Update backend
	if err := c.Bind(backend); err != nil {
		return err
	}
	backend.ID = id

	// Persist updated backend
	if err := backend.Save(reqCtx, tx); err != nil {
		return err
	}

	if err := tx.Commit(reqCtx); err != nil {
		return err
	}

	c.JSON(http.StatusOK, backend)

	return nil
}
