package cluster

import (
	"log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The Gateway accepts bbb cluster requests and dispatches
// it to the cluster nodes.
type Gateway struct {
	state       *State
	handlerFunc HandlerFunc
}

// NewGateway sets up a new cluster router instance.
func NewGateway(state *State) *Gateway {
	return &Gateway{
		state:       state,
		handlerFunc: nilHandler,
	}
}

// Start initializes the router
func (gw *Gateway) Start() {
	log.Println("Starting cluster gateway.")
}

// Register handler
func (gw *Gateway) Register(handler Handler) {
	gw.handlerFunc = handler.Middleware(gw.handlerFunc)
	log.Println("Gateway registered handler:", handler.ID())
}

// Use registers a middleware function
func (gw *Gateway) Use(mware MiddlewareFunc) {
	gw.handlerFunc = mware(gw.handlerFunc)
}

// Dispatch taks a cluster request and starts the middleware
// chain.
func (gw *Gateway) Dispatch(request *bbb.Request) *bbb.Response {
	// Make cluster request and initialize context
	cReq := &Request{Request: request}
	res, err := gw.handlerFunc(cReq)
	if err != nil {
		// TODO: Make generic BBB error response
		return nil
	}
	_ = res
	return nil
}
