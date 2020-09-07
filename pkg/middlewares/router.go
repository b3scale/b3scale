package middleware

import (
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// A Router provides a requets middleware for routing
// requests to backends.
// The routing middleware stack selects backends.
type Router struct {
	state       *cluster.State
	handlerFunc cluster.RouterHandlerFunc
}

// NewRouter creates a new router middleware selecting
// a list of backends from the cluster state
func NewRouter(
	state cluster.State,
	middleware cluster.RouterMiddleware,
) cluster.HandlerFunc {
	return &Router{
		state:       state,
		handlerFunc: routerHandler,
	}
}

// routerHandler is the end of the middleware chain
// it will assert the presence of the "backends"
// slice in the request context.
func routerHandler(req *bbb.Request) (*bbb.Request, error) {
	// TODO: Implement Me
	return req, nil
}
