package cluster

import (
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
		middleware: nilRequestHandler,
	}
}

// nilHandler is an empty handler, that only will result
// in an error when called.
func nilRequestHandler(_req *Request) (Response, error) {
	return nil, fmt.Errorf("end of middleware chain")
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
// chain.
func (gw *Gateway) Dispatch(request *bbb.Request) *bbb.Response {
	// Make cluster request and initialize context
	cReq := &Request{Request: request}
	res, err := gw.middleware(cReq)
	if err != nil {
		// TODO: Make generic BBB error response
		return nil
	}
	_ = res
	return nil
}
