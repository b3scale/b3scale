package cluster

import (
	"context"
	"fmt"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The Router provides a requets middleware for routing
// requests to backends.
// The routing middleware stack selects backends.
type Router struct {
	state      *State
	middleware RouterHandler
}

// NewRouter creates a new router middleware selecting
// a list of backends from the cluster state.
//
// The request will be initialized with a list
// of all backends available in the the cluster state
//
// The middleware chain should only subtract backends.
func NewRouter(state *State) *Router {
	return &Router{
		state:      state,
		middleware: identityHandler,
	}
}

// identityHandler is the end of the middleware chain.
// It will just pass on the backends.
func identityHandler(
	backends []*Backend, _req *bbb.Request,
) ([]*Backend, error) {
	return backends, nil
}

// Use will insert a middleware into the chain
func (r *Router) Use(middleware RouterMiddleware) {
	r.middleware = middleware(r.middleware)
}

// Middleware builds a request middleware
func (r *Router) Middleware() RequestMiddleware {
	return func(next RequestHandler) RequestHandler {
		return func(
			ctx context.Context, req *bbb.Request,
		) (bbb.Response, error) {
			// Filter backends and only accept state active
			backends := make([]*Backend, 0, len(r.state.backends))
			for _, backend := range r.state.backends {
				if backend.State == BackendStateReady {
					backends = append(backends, backend)
				}
			}
			if len(backends) == 0 {
				return nil, fmt.Errorf("no backends available")
			}

			// Add all backends to context and do routing
			backends, err := r.middleware(backends, req)
			if err != nil {
				return nil, err
			}

			// Let other middlewares handle the request
			ctx = ContextWithBackends(ctx, backends)
			return next(ctx, req)
		}
	}
}
