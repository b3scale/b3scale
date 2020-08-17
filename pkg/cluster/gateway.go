package cluster

import (
	"log"
)

// The Gateway accepts bbb cluster requests and dispatches
// it to the cluster nodes.
type Gateway struct {
	controller *Controller
}

// NewGateway sets up a new cluster router instance.
func NewGateway(controller *Controller) *Gateway {
	return &Gateway{controller: controller}
}

// Start initializes the router
func (g *Gateway) Start() {
	log.Println("Starting cluster gateway.")
}

// Dispatch taks a cluster request and forwards it
// to selected number of cluster nodes.
// The responses are then collected, filtered and
// merged.
func (g *Gateway) Dispatch(request *Request) *Response {
	return nil
}
