package api

import (
	"context"

	"github.com/labstack/echo/v4"
)

// ResourceMiddleware is a function returning a HandlerFunction
type ResourceMiddleware func(next ResourceHandler) ResourceHandler

// ResourceHandler is a handler function for API endpoints
type ResourceHandler func(context.Context, *API) error

// Endpoint wrapps a handler function and provides the
// request context and sets the correct APIContext type.
func Endpoint(handler ResourceHandler) echo.HandlerFunc {
	return func(c echo.Context) error {
		api := c.(*API)
		ctx := api.Ctx()
		return handler(ctx, api)
	}
}

// Resource is a restful handler group
type Resource struct {
	List    ResourceHandler
	Show    ResourceHandler
	Create  ResourceHandler
	Update  ResourceHandler
	Destroy ResourceHandler
}

// Mount registers the endpoints in the group
func (r *Resource) Mount(
	root *echo.Group,
	prefix string,
	middlewares ...echo.MiddlewareFunc,
) {
	g := root.Group(prefix, middlewares...)
	if r.List != nil {
		g.GET("", Endpoint(r.List))
	}
	if r.Create != nil {
		g.POST("", Endpoint(r.Create))
	}
	if r.Show != nil {
		g.GET("/:id", Endpoint(r.Show))
	}
	if r.Update != nil {
		g.PATCH("/:id", Endpoint(r.Update))
	}
	if r.Destroy != nil {
		g.DELETE("/:id", Endpoint(r.Destroy))
	}
}
