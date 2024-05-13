package cluster

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/bbb"
)

// Errors
var (
	// ErrNoBackendInContext will be returned when no backends
	// could be associated with the request.
	ErrNoBackendInContext = errors.New("no backend in context")

	// ErrNoFrontendInContext will be returned when no frontend
	// is associated with the request.
	ErrNoFrontendInContext = errors.New("no frontend in context")

	// ErrBackendNotReady will only occur when the routing
	// selected a backend that can not accept any requests
	ErrBackendNotReady = errors.New("backend not ready")
)

// GatewayOptions have flags for customizing the gateway behavior.
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
	gw.middleware = gw.unhandledRequestHandler(ctrl)
	return gw
}

// The unhandledRequestHandler marks the end of the
// middleware chain and is the default handler for requests.
func (gw *Gateway) unhandledRequestHandler(ctrl *Controller) RequestHandler {
	return func(
		ctx context.Context, req *bbb.Request,
	) (bbb.Response, error) {
		// We could not handle the request.
		return nil, fmt.Errorf("unknown resource: %s", req.Resource)
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
		res = &bbb.XMLResponse{
			Returncode: bbb.RetFailed,
			MessageKey: "b3scaleGatewayError",
			Message:    fmt.Sprintf("%s", err),
		}
	}
	return res
}
