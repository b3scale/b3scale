package v1

import (
	"context"

	"github.com/labstack/echo/v4"
)

// APIEndpointMiddleware is a function returning a HandlerFunction
type APIEndpointMiddleware func(next APIEndpointHandler) APIEndpointHandler

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
		g.GET("", APIEndpoint(r.List))
	}
	if r.Create != nil {
		g.POST("", APIEndpoint(r.Create))
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
