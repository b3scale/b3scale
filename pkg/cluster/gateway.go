package cluster

import (
	"context"
	"fmt"
	"log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The Gateway accepts bbb cluster requests and dispatches
// it to the cluster nodes.
type Gateway struct {
	state      *State
	middleware RequestHandler
}

// NewGateway sets up a new cluster router instance.
func NewGateway(state *State) *Gateway {
	return &Gateway{
		state:      state,
		middleware: dispatchBackendHandler,
	}
}

// The dispatchBackendHandler marks the end of the
// middleware chain and is the default handler for requests.
// It expects the presence of a "backend" in the current
// context. Otherwise it will fail.
func dispatchBackendHandler(
	ctx context.Context, req *bbb.Request,
) (bbb.Response, error) {
	// Get backend by id
	backend := BackendFromContext(ctx)
	if backend == nil {
		return nil, fmt.Errorf("no backend in context")
	}

	// Check if the backend is ready to accept requests:
	if backend.State != BackendStateReady {
		return nil, fmt.Errorf("backend not ready")
		// This should however not happen!
	}

	// Dispatch API resources
	switch req.Resource {
	case bbb.ResJoin:
		return backend.Join(req)
	case bbb.ResCreate:
		return backend.Create(req)
	case bbb.ResIsMeetingRunning:
		return backend.IsMeetingRunning(req)
	case bbb.ResEnd:
		return backend.End(req)
	case bbb.ResGetMeetingInfo:
		return backend.GetMeetingInfo(req)
	case bbb.ResGetMeetings:
		return backend.GetMeetings(req)
	case bbb.ResGetRecordings:
		return backend.GetRecordings(req)
	case bbb.ResPublishRecordings:
		return backend.PublishRecordings(req)
	case bbb.ResDeleteRecordings:
		return backend.DeleteRecordings(req)
	case bbb.ResUpdateRecordings:
		return backend.UpdateRecordings(req)
	case bbb.ResGetDefaultConfigXML:
		return backend.GetDefaultConfigXML(req)
	case bbb.ResSetConfigXML:
		return backend.SetConfigXML(req)
	case bbb.ResGetRecordingTextTracks:
		return backend.GetRecordingTextTracks(req)
	case bbb.ResPutRecordingTextTrack:
		return backend.PutRecordingTextTrack(req)
	}
	// We could not dispatch this
	return nil, fmt.Errorf(
		"unknown resource: %s", req.Resource)
}

// Start initializes the router
func (gw *Gateway) Start() {
	log.Println("Starting cluster gateway.")
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

	// Make cluster request and initialize context
	res, err := gw.middleware(ctx, req)
	if err != nil {
		// We encode our error as a BBB error response
		return &bbb.XMLResponse{
			Returncode: "FAILED",
			MessageKey: "b3scale_gateway_error",
			Message:    fmt.Sprintf("%s", err),
		}
	}
	return res
}
