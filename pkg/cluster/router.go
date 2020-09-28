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
		middleware: selectDiscardHandler,
	}
}

// As a final step in routing, make sure that there
// is only a single backend left in the list of
// potential backends for some resources.
//
// This pretty much applies to all state mutating
// API resources like join or create.
//
// We use the selectDiscardHandler as the end of our
// middleware chain.
func selectDiscardHandler(
	backends []*Backend, req *bbb.Request,
) ([]*Backend, error) {
	res := req.Resource
	switch res {
	case bbb.ResJoin:
		return selectFirst(backends), nil
	case bbb.ResCreate:
		return selectFirst(backends), nil
	case bbb.ResIsMeetingRunning:
		return selectFirst(backends), nil
	case bbb.ResEnd:
		return selectFirst(backends), nil
	case bbb.ResGetMeetingInfo:
		return selectFirst(backends), nil
	case bbb.ResGetMeetings:
		return backends, nil
	case bbb.ResGetRecordings:
		return backends, nil
	case bbb.ResPublishRecordings:
		return backends, nil
	case bbb.ResDeleteRecordings:
		return backends, nil
	case bbb.ResUpdateRecordings:
		return selectFirst(backends), nil
	case bbb.ResGetDefaultConfigXML:
		return selectFirst(backends), nil
	case bbb.ResSetConfigXML:
		return selectFirst(backends), nil
	case bbb.ResGetRecordingTextTracks:
		return selectFirst(backends), nil
	case bbb.ResPutRecordingTextTrack:
		return selectFirst(backends), nil
	}

	return nil, fmt.Errorf(
		"unknown api resource for backend select: %s", res)
}

// Keep only first backend
func selectFirst(backends []*Backend) []*Backend {
	// The following slice operation with empty slices
	if len(backends) == 0 {
		return backends
	}
	return backends[:1]
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
