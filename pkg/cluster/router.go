package cluster

import (
	"context"
	"fmt"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// The Router provides a requets middleware for routing
// requests to backends.
// The routing middleware stack selects backends.
type Router struct {
	ctrl       *Controller
	middleware RouterHandler
}

// NewRouter creates a new router middleware selecting
// a list of backends from the cluster state.
//
// The request will be initialized with a list
// of all backends available in the the cluster state
//
// The middleware chain should only subtract backends.
func NewRouter(ctrl *Controller) *Router {
	return &Router{
		ctrl:       ctrl,
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
// EDIT: For now we are just using a single
//       backend and let the backend implementation
//       hit the database instead of the actual backends.
//
// We use the selectDiscardHandler as the end of our
// middleware chain.
func selectDiscardHandler(
	backends []*Backend, req *bbb.Request,
) ([]*Backend, error) {
	res := req.Resource
	switch res {
	case bbb.ResourceJoin:
		return selectFirst(backends), nil
	case bbb.ResourceCreate:
		return selectFirst(
			discardShutdown(backends)), nil
	case bbb.ResourceIsMeetingRunning:
		return selectFirst(backends), nil
	case bbb.ResourceEnd:
		return selectFirst(backends), nil
	case bbb.ResourceGetMeetingInfo:
		return selectFirst(backends), nil
	case bbb.ResourceGetMeetings:
		return selectFirst(backends), nil
	case bbb.ResourceGetRecordings:
		return selectFirst(backends), nil
	case bbb.ResourcePublishRecordings:
		return selectFirst(backends), nil
	case bbb.ResourceDeleteRecordings:
		return selectFirst(backends), nil
	case bbb.ResourceUpdateRecordings:
		return selectFirst(backends), nil
	case bbb.ResourceGetDefaultConfigXML:
		return selectFirst(backends), nil
	case bbb.ResourceSetConfigXML:
		return selectFirst(backends), nil
	case bbb.ResourceGetRecordingTextTracks:
		return selectFirst(backends), nil
	case bbb.ResourcePutRecordingTextTrack:
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

// Keep only backends with admin state ready
func discardShutdown(backends []*Backend) []*Backend {
	filtered := make([]*Backend, 0, len(backends))
	for _, b := range backends {
		if b.state.AdminState != "ready" {
			continue
		}
		filtered = append(filtered, b)
	}
	return filtered
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
			// Filter backends and only accept state active,
			// and where the noded is active on the host.
			backends, err := r.ctrl.GetBackends(store.Q().
				Where(`
					backends.id NOT IN (
						 SELECT id
						   FROM backends_node_offline
							FOR UPDATE SKIP LOCKED)`).
				Where("node_state = ?", "ready"))
			if err != nil {
				return nil, err
			}
			if len(backends) == 0 {
				return nil, fmt.Errorf("no backends available")
			}

			// Apply routing middleware to backends for a BBB request
			backends, err = r.middleware(backends, req)
			if err != nil {
				return nil, err
			}

			// Let other middlewares handle the request
			ctx = ContextWithBackends(ctx, backends)
			return next(ctx, req)
		}
	}
}
