package cluster

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/store"
)

// Routing errors
var (
	// ErrNoBackendForMeeting indicates, that there is a backend
	// expected to be associated with a meeting, yet the meeting
	// is unknown to the cluster.
	ErrNoBackendForMeeting = errors.New("no backends associated with meeting")

	// ErrNoBackendAvailable indicates that there is no backend
	// available for creating a meeting.
	ErrNoBackendAvailable = errors.New("no free backend available for meeting")

	// ErrMeetingIDMissing indicates that there is a meetingID
	// expected to be in the requests params, but it is missing.
	ErrMeetingIDMissing = errors.New("meetingID missing from request")
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
		middleware: nilHandler,
	}
}

// The nil handler is the end of the middleware chain.
// It will just return the backends as they are.
func nilHandler(
	ctx context.Context,
	backends []*Backend,
	req *bbb.Request,
) ([]*Backend, error) {
	return backends, nil
}

// Use will insert a middleware into the chain
func (r *Router) Use(middleware RouterMiddleware) {
	r.middleware = middleware(r.middleware)
}

// SelectBackend will apply the routing middleware
// chain to a given request with all ready nodes in
// the cluster where the admin state is also ready.
// Selecting a backend will fail if no backends are available
// as routing targets.
func (r *Router) SelectBackend(
	ctx context.Context, req *bbb.Request,
) (*Backend, error) {
	// Filter backends and only accept state active,
	// and where the node agent is active on the host.
	// Also we exclude stopped nodes.
	deadline := time.Now().UTC().Add(-5 * time.Second)
	backends, err := GetBackends(ctx, store.Q().
		Where("agent_heartbeat >= ?", deadline).
		Where("admin_state = ?", "ready").
		Where("node_state = ?", "ready"))
	if err != nil {
		return nil, err
	}
	backends, err = r.middleware(ctx, backends, req)
	if err != nil {
		return nil, err
	}
	if len(backends) == 0 {
		return nil, ErrNoBackendAvailable
	}

	// Use first backend
	return backends[0], nil
}

// LookupBackend will retrieve a backend or will fail
// if the backend could not be found. Primary identifier
// is the MeetingID of the request.
// When no backend is found, this will not fail, however
// the backend will be nil and this case needs to be handled.
func (r *Router) LookupBackend(
	ctx context.Context,
	req *bbb.Request,
) (*Backend, error) {
	// Get meeting id from params. If none is present,
	// there is nothing to do for us here.
	meetingID, ok := req.Params.MeetingID()
	if !ok {
		return nil, ErrMeetingIDMissing
	}
	log.Debug().
		Str("meetingID", meetingID).
		Msg("lookupBackendForRequest")

	// Lookup backend for meeting in cluster, use backend
	// if there is one associated.
	backend, err := GetBackend(ctx, store.Q().
		Join("meetings ON meetings.backend_id = backends.id").
		Where("meetings.id = ?", meetingID))
	if err != nil {
		return nil, err
	}
	if backend == nil {
		log.Debug().
			Str("meetingID", meetingID).
			Msg("no backend for meetingID")
		return nil, nil
	}
	log.Debug().
		Str("meetingID", meetingID).
		Str("backend", backend.state.Backend.Host).
		Msg("found backend for meetingID")
	return backend, nil
}

// LookupBackendForRecordID uses the recordID to identify
// a backend via the recordings state table.
func (r *Router) LookupBackendForRecordID(
	ctx context.Context,
	recordID string,
) (*Backend, error) {
	log.Debug().
		Str("recordID", recordID).
		Msg("lookupBackendForRecordID")

	// Lookup backend for meeting in cluster, use backend
	// if there is one associated.
	backend, err := GetBackend(ctx, store.Q().
		Join("recordings ON recordings.backend_id = backends.id").
		Where("recordings.record_id = ?", recordID))
	if err != nil {
		return nil, err
	}
	if backend == nil {
		log.Warn().
			Str("recordID", recordID).
			Msg("no backend for recordID")
		return nil, nil
	}
	log.Debug().
		Str("recordID", recordID).
		Str("backend", backend.state.Backend.Host).
		Msg("found backend for recordID")
	return backend, nil
}
