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
	ref := ctx.FilterAccountRef()

	// Begin TX
	tx, err := store.ConnectionFromContext(reqCtx).Begin(reqCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
	}
	defer tx.Rollback(reqCtx)

	// Begin Query
	q := store.Q()
	if ref != nil {
		q.Where("account_ref = ?", *ref)
	}

	backends, err := store.GetBackendStates(reqCtx, tx, q)
	c.JSON(http.StatusOK, backends)

	return nil
}

// BackendCreate will add a new frontend to the cluster.
// ! requires: `admin`
func BackendCreate(c echo.Context) error {
	return nil
}

// BackendRetrieve will retrieve a single frontend
// identified by ID.
// ! requires: `admin`
func BackendRetrieve(c echo.Context) error {
	return nil
}

// BackendDestroy will remove a frontend from the cluster.
// The frontend is identified by ID.
// ! requires: `admin`
func BackendDestroy(c echo.Context) error {
	return nil
}

// BackendUpdate will update the frontend with values
// provided by the request. Only keys provided will
// be updated.
// ! requires: `admin`
func BackendUpdate(c echo.Context) error {
	return nil
}
