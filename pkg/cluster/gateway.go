package cluster

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The Gateway accepts bbb cluster requests and dispatches
// it to the cluster nodes.
type Gateway struct {
	middleware RequestHandler
	ctrl       *Controller
}

// NewGateway sets up a new cluster router instance.
func NewGateway(ctrl *Controller) *Gateway {
	return &Gateway{
		ctrl:       ctrl,
		middleware: dispatchBackendHandler(ctrl),
	}
}

// The dispatchBackendHandler marks the end of the
// middleware chain and is the default handler for requests.
// It expects the presence of a "backend" in the current
// context. Otherwise it will fail.
func dispatchBackendHandler(ctrl *Controller) RequestHandler {
	return func(
		ctx context.Context, req *bbb.Request,
	) (bbb.Response, error) {
		// Get backend and frontend
		backends := BackendsFromContext(ctx)
		if len(backends) == 0 {
			return nil, fmt.Errorf("no backend in context")
		}
		backend := backends[0]

		frontend := FrontendFromContext(ctx)
		if frontend == nil {
			return nil, fmt.Errorf("no frontend in context")
		}

		// Check if the backend is ready to accept requests:
		if backend.state.NodeState != BackendStateReady {
			return nil, fmt.Errorf("backend not ready")
			// This should however not happen!
		}

		// Make sure the meeting is associated with the backend
		if meetingID, ok := req.Params.MeetingID(); ok {
			meeting, err := ctrl.GetMeetingStateByID(meetingID)
			if err != nil {
				log.Error().
					Err(err).
					Str("meetingID", meetingID).
					Msg("GetMeetingStateByID")
			} else {
				if meeting != nil {
					// Assign to backend
					if err := meeting.SetBackendID(backend.ID()); err != nil {
						return nil, err
					}
					// Assign frontend if not present
					if err := meeting.BindFrontendID(frontend.state.ID); err != nil {
						return nil, err
					}
				}
			}
		}

		// Dispatch API resources
		switch req.Resource {
		case bbb.ResourceJoin:
			return backend.Join(req)
		case bbb.ResourceCreate:
			return backend.Create(req)
		case bbb.ResourceIsMeetingRunning:
			return backend.IsMeetingRunning(req)
		case bbb.ResourceEnd:
			return backend.End(req)
		case bbb.ResourceGetMeetingInfo:
			return backend.GetMeetingInfo(req)
		case bbb.ResourceGetMeetings:
			return backend.GetMeetings(req)
		case bbb.ResourceGetRecordings:
			return backend.GetRecordings(req)
		case bbb.ResourcePublishRecordings:
			return backend.PublishRecordings(req)
		case bbb.ResourceDeleteRecordings:
			return backend.DeleteRecordings(req)
		case bbb.ResourceUpdateRecordings:
			return backend.UpdateRecordings(req)
		case bbb.ResourceGetDefaultConfigXML:
			return backend.GetDefaultConfigXML(req)
		case bbb.ResourceSetConfigXML:
			return backend.SetConfigXML(req)
		case bbb.ResourceGetRecordingTextTracks:
			return backend.GetRecordingTextTracks(req)
		case bbb.ResourcePutRecordingTextTrack:
			return backend.PutRecordingTextTrack(req)
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
func (gw *Gateway) Dispatch(ctx context.Context, req *bbb.Request) bbb.Response {
	// Trigger backed jobs
	go gw.ctrl.StartBackground()

	// Make cluster request and initialize context
	res, err := gw.middleware(ctx, req)
	if err != nil {
		// Log the error
		log.Error().
			Err(err).
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
