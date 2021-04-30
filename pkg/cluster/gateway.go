package cluster

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// Errors
var (
	// ErrNoBackendInContext will be returned when no backends
	// could be associated with the request.
	ErrNoBackendInContext = errors.New("no backend in context")

	// ErrNoFrontendInContext will be returned when no frontend
	// is associated with the request.
	ErrNoFrontendInContext = errors.New("no fontend in context")

	// ErrBackendNotReady will only occure when the routing
	// selected a backend that can not accept any requests
	ErrBackendNotReady = errors.New("backend not ready")
)

// GatewayOptions have flags for customizing the gateway behaviour.
type GatewayOptions struct {
}

// The Gateway accepts bbb cluster requests and dispatches
// it to the cluster nodes.
type Gateway struct {
	opts       *GatewayOptions
	middleware RequestHandler
	ctrl       *Controller
}

// NewGateway sets up a new cluster router instance.
func NewGateway(ctrl *Controller, opts *GatewayOptions) *Gateway {
	gw := &Gateway{
		ctrl: ctrl,
		opts: opts,
	}
	gw.middleware = gw.dispatchBackendHandler(ctrl)
	return gw
}

// The dispatchBackendHandler marks the end of the
// middleware chain and is the default handler for requests.
// It expects the presence of a "backend" in the current
// context. Otherwise it will fail.
func (gw *Gateway) dispatchBackendHandler(ctrl *Controller) RequestHandler {
	return func(
		ctx context.Context, req *bbb.Request,
	) (bbb.Response, error) {
		// Get backend and frontend and dispatch to the first
		// backend in the set.
		backends := BackendsFromContext(ctx)
		if len(backends) == 0 {
			return nil, ErrNoBackendInContext
		}
		backend := backends[0]

		frontend := FrontendFromContext(ctx)
		if frontend == nil {
			return nil, ErrNoFrontendInContext
		}

		// Check if the backend is ready to accept requests
		if backend.state.NodeState != BackendStateReady {
			return nil, ErrBackendNotReady
		}

		// Make sure the meeting is associated with the backend
		if meetingID, ok := req.Params.MeetingID(); ok {
			// Next, we have to access our shared state
			tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
			if err != nil {
				return nil, err
			}
			defer tx.Rollback(ctx)

			meeting, err := store.GetMeetingStateByID(ctx, tx, meetingID)
			if err != nil {
				log.Error().
					Err(err).
					Str("meetingID", meetingID).
					Msg("GetMeetingStateByID")
			} else {
				if meeting != nil {
					// Assign to backend
					if err := meeting.SetBackendID(ctx, tx, backend.ID()); err != nil {
						return nil, err
					}
					// Assign frontend if not present
					if err := meeting.BindFrontendID(ctx, tx, frontend.state.ID); err != nil {
						return nil, err
					}
				}
			}

			// Persist changes and close transaction
			if err := tx.Commit(ctx); err != nil {
				return nil, err
			}
		}

		// Dispatch API resources
		switch req.Resource {
		case bbb.ResourceIndex:
			return backend.Version(req)
		case bbb.ResourceJoin:
			if gw.opts.IsReverseProxyEnabled {
				return backend.JoinProxy(ctx, req)
			}
			return backend.Join(ctx, req)
		case bbb.ResourceCreate:
			return backend.Create(ctx, req)
		case bbb.ResourceIsMeetingRunning:
			return backend.IsMeetingRunning(ctx, req)
		case bbb.ResourceEnd:
			return backend.End(ctx, req)
		case bbb.ResourceGetMeetingInfo:
			return backend.GetMeetingInfo(ctx, req)
		case bbb.ResourceGetMeetings:
			return backend.GetMeetings(ctx, req)
		case bbb.ResourceGetRecordings:
			return backend.GetRecordings(ctx, req)
		case bbb.ResourcePublishRecordings:
			return backend.PublishRecordings(ctx, req)
		case bbb.ResourceDeleteRecordings:
			return backend.DeleteRecordings(ctx, req)
		case bbb.ResourceUpdateRecordings:
			return backend.UpdateRecordings(ctx, req)
		case bbb.ResourceGetDefaultConfigXML:
			return backend.GetDefaultConfigXML(ctx, req)
		case bbb.ResourceSetConfigXML:
			return backend.SetConfigXML(ctx, req)
		case bbb.ResourceGetRecordingTextTracks:
			return backend.GetRecordingTextTracks(ctx, req)
		case bbb.ResourcePutRecordingTextTrack:
			return backend.PutRecordingTextTrack(ctx, req)
		}

		// We could not dispatch this
		return nil, fmt.Errorf(
			"unknown resource: %s", req.Resource)
	}
}

// Use registers a middleware function
func (gw *Gateway) Use(middleware RequestMiddleware) {
	gw.middleware = middleware(gw.middleware)
}

// Dispatch taks a cluster request and starts the middleware
// chain. We will always return a bbb response.
// Any error occoring during routing or dispatching will be
// encoded as an BBB XML Response.
func (gw *Gateway) Dispatch(
	ctx context.Context,
	conn *pgxpool.Conn,
	req *bbb.Request,
) bbb.Response {
	// Trigger backed jobs
	go gw.ctrl.StartBackground()

	// Let the middleware chain handle the request
	res, err := gw.middleware(ctx, req)
	if err != nil {
		be := BackendFromContext(ctx)
		fe := FrontendFromContext(ctx)
		// Log the error
		log.Error().
			Err(err).
			Str("backend", fmt.Sprintf("%v", be)).
			Str("frontend", fmt.Sprintf("%v", fe)).
			Msg("gateway error")
		// We encode our error as a BBB error response
		return &bbb.XMLResponse{
			Returncode: "FAILED",
			MessageKey: "b3scale_gateway_error",
			Message:    fmt.Sprintf("%s", err),
		}
	}
	return res
}
