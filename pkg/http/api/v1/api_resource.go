package v1

import (
	"context"

	"github.com/labstack/echo/v4"
)

// APIEndpointHandler is a handler function for API endpoints
type APIEndpointHandler func(context.Context, *APIContext) error

// APIEndpoint wrapps a handler function and provides the
// request context and sets the correct APIContext type.
func APIEndpoint(handler APIEndpointHandler) echo.HandlerFunc {
	return func(c echo.Context) error {
		api := c.(*APIContext)
		ctx := api.Ctx()
		return handler(ctx, api)
	}
}

// APIResource is a restful handler group
type APIResource struct {
	List    APIEndpointHandler
	Show    APIEndpointHandler
	Create  APIEndpointHandler
	Update  APIEndpointHandler
	Destroy APIEndpointHandler
}

// Mount registers the endpoints in the group
func (r *APIResource) Mount(
	root *echo.Group,
	prefix string,
	middlewares ...echo.MiddlewareFunc,
) {
	g := root.Group(prefix, middlewares...)
	if r.List != nil {
		g.GET("/", APIEndpoint(r.List))
	}
	if r.Create != nil {
		g.POST("/", APIEndpoint(r.Create))
	}
	if r.Show != nil {
		g.GET("/:id", APIEndpoint(r.Show))
	}
	if r.Update != nil {
		g.PATCH("/:id", APIEndpoint(r.Update))
	}
	if r.Destroy != nil {
		g.DELETE("/:id", APIEndpoint(r.Destroy))
	}
}

// APIResource makes a new resource group
/*
func APIResource(g *echo.Group) *APIResourceGroup {
	return &APIResourceGroup{g}
}

// APIResourceGroup is a restful endpoint group
type APIResourceGroup struct {
	*echo.Group
}

// List registers an API endpoint handler as the resource index
func (res *APIResourceGroup) List(
	handler APIEndpointHandler,
) *APIResourceGroup {
	res.GET("/", APIEndpoint(handler))
	return res
}

// Create registers an API endpoint for posting a new resource
func (res *APIResourceGroup) Create(
	handler APIEndpointHandler,
) *APIResourceGroup {
	res.POST("/", APIEndpoint(handler))
	return res
}

// Show retrieves an API resource
func (res *APIResourceGroup) Show(
	handler APIEndpointHandler,
) *APIResourceGroup {
	res.GET("/:id", APIEndpoint(handler))
	return res
}

// Update registers a patch endpoint
func (res *APIResourceGroup) Update(
	handler APIEndpointHandler,
) *APIResourceGroup {
	res.PATCH("/:id", APIEndpoint(handler))
	return res
}

// Destroy registers a DELETE endpoint
func (res *APIResourceGroup) Destroy(
	handler APIEndpointHandler,
) *APIResourceGroup {
	res.DELETE("/:id", APIEndpoint(handler))
	return res
}
*/
