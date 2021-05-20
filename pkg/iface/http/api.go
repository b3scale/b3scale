package http

/*
B3Scale API v1

Administrative API for B3Scale. See /docs/rest_api.md for
details.
*/

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// API is the v1 implementation of the b3scale
// administrative API.
type API struct{}

// InitAPI sets up a group with authentication
// for a restful management interface.
func InitAPI(e *echo.Echo) error {
	// Initialize JWT middleware and authorization middleware

	// Initialize API
	api := &API{}

	// Register routes
	a := e.Group("/api/v1")

	// Frontends
	a.GET("/frontends", api.FrontendsList)
	a.POST("/frontends", api.FrontendCreate)
	a.GET("/frontends/:id", api.FrontendRetrieve)
	a.DELETE("/frontends/:id", api.FrontendDestroy)
	a.PATCH("/frontends/:id", api.FrontendUpdate)

	// Backends
	a.GET("/backends", api.BackendsList)
	a.POST("/backends", api.BackendCreate)
	a.GET("/backends/:id", api.BackendRetrieve)
	a.DELETE("/backends/:id", api.BackendDestroy)
	a.PATCH("/backends/:id", api.BackendUpdate)

	return nil
}

// FrontendsList will list all frontends known
// to the cluster or within the user scope.
func (a *API) FrontendsList(c echo.Context) error {
	return nil
}

// FrontendCreate will add a new frontend to the cluster.
func (a *API) FrontendCreate(c echo.Context) error {
	return nil
}

// FrontendRetrieve will retrieve a single frontend
// identified by ID.
func (a *API) FrontendRetrieve(c echo.Context) error {
	return nil
}

// FrontendDestroy will remove a frontend from the cluster.
// The frontend is identified by ID.
func (a *API) FrontendDestroy(c echo.Context) error {
	return nil
}

// FrontendUpdate will update the frontend with values
// provided by the request. Only keys provided will
// be updated.
func (a *API) FrontendUpdate(c echo.Context) error {
	return nil
}

// BackendsList will list all frontends known
// to the cluster or within the user scope.
func (a *API) BackendsList(c echo.Context) error {
	return nil
}

// BackendCreate will add a new frontend to the cluster.
func (a *API) BackendCreate(c echo.Context) error {
	return nil
}

// BackendRetrieve will retrieve a single frontend
// identified by ID.
func (a *API) BackendRetrieve(c echo.Context) error {
	return nil
}

// BackendDestroy will remove a frontend from the cluster.
// The frontend is identified by ID.
func (a *API) BackendDestroy(c echo.Context) error {
	return nil
}

// BackendUpdate will update the frontend with values
// provided by the request. Only keys provided will
// be updated.
func (a *API) BackendUpdate(c echo.Context) error {
	return nil
}
