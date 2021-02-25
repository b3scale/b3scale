package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

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
// We use the selectDiscardHandler as the end of our
// middleware chain.
func selectDiscardHandler(
	backends []*Backend, req *bbb.Request,
) ([]*Backend, error) {
	res := req.Resource
	switch res {
	case bbb.ResourceIndex:
		return selectFirst(backends), nil
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

// Lookup middleware for retriving an already associated
// backend for a given meeting.
func (r *Router) lookupBackendForRequest(
	ctx context.Context,
	req *bbb.Request,
) (*Backend, error) {
	// Get meeting id from params. If none is present,
	// there is nothing to do for us here.
	meetingID, ok := req.Params.MeetingID()
	if !ok {
		return nil, nil
	}

	log.Debug().
		Str("meetingID", meetingID).
		Msg("lookupBackendForRequest")

	// Lookup backend for meeting in cluster, use backend
	// if there is one associated - otherwise return
	// all possible backends.
	backend, err := r.ctrl.GetBackend(ctx, store.Q().
		Join("meetings ON meetings.backend_id = backends.id").
		Where("meetings.id = ?", meetingID))
	if err != nil {
		return nil, err
	}
	if backend == nil {
		// No specific backend was associated with the ID
		return nil, nil
	}

	// Depending on the request we need to check if
	// the backend can be used for accepting the request.
	// To do so, we apply the routing middleware chain
	// to the backend and see if it is included in the result set.
	res, err := r.middleware([]*Backend{backend}, req)
	if err != nil {
		return nil, err
	}
	if !r.isBackendAvailable(backend, res) {
		// The backend associated with the meeting
		// can not be used for this request. We need to
		// relocate and will destroy the association
		// with the backend.
		if err := r.ctrl.DeleteMeetingStateByID(ctx, meetingID); err != nil {
			log.Error().
				Err(err).
				Msg("failed removing meeting state")
		}
		return nil, nil
	}

	// Okay looks like this backend is a good candidate.
	return backend, nil
}

func (r *Router) isBackendAvailable(
	backend *Backend,
	backends []*Backend,
) bool {
	for _, b := range backends {
		if b.ID() == backend.ID() {
			return true
		}
	}

	// Emit a warning
	log.Warn().
		Str("backend", backend.Host()).
		Str("backendID", backend.ID()).
		Msg("requested backend is no longer available " +
			"as selectable routing target. " +
			"Reassigning meetig.")

	return false
}

// Middleware builds a request middleware
func (r *Router) Middleware() RequestMiddleware {
	return func(next RequestHandler) RequestHandler {
		// Do routing
		return func(
			ctx context.Context, req *bbb.Request,
		) (bbb.Response, error) {
			// Filter backends and only accept state active,
			// and where the noded is active on the host.
			// Also we exclude stopped nodes.
			deadline := time.Now().UTC().Add(-5 * time.Second)
			backends, err := r.ctrl.GetBackends(ctx, store.Q().
				Where("agent_heartbeat >= ?", deadline).
				Where("node_state = ?", "ready"))
			if err != nil {
				return nil, err
			}
			if len(backends) == 0 {
				return nil, fmt.Errorf("no backends available")
			}

			// Try to lookup meeting for the incoming request
			backend, err := r.lookupBackendForRequest(ctx, req)
			if backend != nil {
				log.Debug().
					Str("backendID", backend.ID()).
					Msg("found backend for meeting id")

				// We found a backend! If it is still available, we skip
				// the router middleware chain and invoke the next request
				// middleware with this backend.
				if r.isBackendAvailable(backend, backends) {
					ctx = ContextWithBackends(ctx, []*Backend{backend})
					return next(ctx, req)
				}
			} else {
				// Router only path
				log.Debug().
					Msg("no backend found for meeting... " +
						"applying routing middlewares")
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
