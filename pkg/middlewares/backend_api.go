package middleware

import (
	"fmt"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// BackendMiddleware Handler will invoke API calls
// on backends. The backend is selected
// by popping a backend label.

// NewBackendMiddleware creates the request handler func
func NewBackendMiddleware() cluster.MiddlewareFunc {
	return func(_next cluster.HandlerFunc) cluster.HandlerFunc {
		return dispatch
	}
}

// Request Handler
func dispatch(req *cluster.Request) (*cluster.Response, error) {
	// Get backend by id
	cBackend, ok := req.Context.Load("backend")
	if !ok {
		return nil, fmt.Errorf("no backend in context")
	}

	// The backend implements the BBB API
	backend := cBackend.(bbb.API)

	// Dispatch API Request
	var res bbb.Response
	switch res.Resource {
	case bbb.APIJoin:
		return backend.Join(req.Request)
	case bbb.APICreate:
		return backend.Create(req.Request)
	case bbb.APIIsMeetingRunning:
		return backend.IsMeetingRunning(req.Request)
	case bbb.APIEnd:
		return backend.End(req.Request)

	}

	return nil, fmt.Errorf("implement me")
}
