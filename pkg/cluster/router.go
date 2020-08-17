package cluster

// BBB Cluster Router

import (
	"log"
)

// The Router accepts bbb cluster requests and dispatches
// it to the cluster nodes.
type Router struct {
	controller *Controller
}

// NewRouter sets up a new cluster router instance.
func NewRouter(controller *Controller) *Router {
	return &Router{controller: controller}
}

// Start initializes the router
func (r *Router) Start() {
	log.Println("Starting router.")
}

// Dispatch taks a cluster request and forwards it
// to selected number of cluster nodes.
// The responses are then collected, filtered and
// merged.
func (r *Router) Dispatch(request *Request) *Response {
	return nil
}
