package cluster

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
// It will just pass on the request.
func identityHandler(req *Request) (*Request, error) {
	return req, nil
}

// Use will insert a middleware into the chain
func (r *Router) Use(middleware RouterMiddleware) {
	r.middleware = middleware(r.middleware)
}

// Middleware builds a request middleware
func (r *Router) Middleware(
	next RequestHandler,
) RequestHandler {
	return func(req *Request) (Response, error) {
		// Add all backends to context
		req.Context = ContextWithBackends(
			req.Context, r.state.backends)

		// Do routing
		req, err := r.middleware(req)
		if err != nil {
			return nil, err
		}

		// Let other middlewares handle the request
		return next(req)
	}
}
