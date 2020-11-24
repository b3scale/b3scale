package cluster

import (
	"context"
	"fmt"
	"log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// BackendStates: The state of the cluster backend.
const (
	BackendStateInit    = "init"
	BackendStateReady   = "ready"
	BackendStateError   = "error"
	BackendStateStopped = "stopped"
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
		// Get backend by id
		backend := BackendFromContext(ctx)
		if backend == nil {
			return nil, fmt.Errorf("no backend in context")
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
				log.Println(err)
			} else {
				if meeting != nil {
					if err := meeting.SetBackendID(backend.ID()); err != nil {
						log.Println(err)
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
func (gw *Gateway) Dispatch(req *bbb.Request) bbb.Response {
	// Make initial context
	ctx := context.Background()

	// Trigger backed jobs
	go gw.ctrl.StartBackground()

	// Make cluster request and initialize context
	res, err := gw.middleware(ctx, req)
	if err != nil {
		// Log the error
		log.Println("gateway error:", err)
		// We encode our error as a BBB error response
		return &bbb.XMLResponse{
			Returncode: "FAILED",
			MessageKey: "b3scale_gateway_error",
			Message:    fmt.Sprintf("%s", err),
		}
	}
	return res
}
