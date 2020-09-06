package cluster

import (
	"log"
)

// The Gateway accepts bbb cluster requests and dispatches
// it to the cluster nodes.
type Gateway struct {
	state      *State
	middleware HandlerFunc
}

// NewGateway sets up a new cluster router instance.
func NewGateway(state *State) *Gateway {
	return &Gateway{
		state:      state,
		middleware: NoneHandler,
	}
}

// Start initializes the router
func (gw *Gateway) Start() {
	log.Println("Starting cluster gateway.")
}

// Register handler
func (gw *Gateway) Register(handler Handler) {
	gw.middleware = handler.Middleware(gw.middleware)
}

// Dispatch taks a cluster request and forwards it
// to selected number of cluster nodes.
// The responses are then collected, filtered and
// merged.
func (gw *Gateway) Dispatch(request *Request) *Response {
	return nil
}
