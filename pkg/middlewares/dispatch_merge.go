package middleware

// Dispatch / Merge Middleware
// Forwards the request to the selected backends
// and merges the response.

import (
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

type dispatchResult struct {
	response *cluster.Response
	error    error
}

// NewDispatchMerge creates the request middleware
// function, dispatching the request to nodes and
// collects responses.
func NewDispatchMerge() cluster.RequestMiddleware {
	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return func(req *cluster.Request) (cluster.Response, error) {
			// Fanout request, collect responses and
			// call next middleware with each response.
			return dispatchMerge(next, req)
		}
	}
}

// Actual dispatch and merge of the request.
func dispatchMerge(
	next cluster.RequestHandler,
	req *cluster.Request,
) (cluster.Response, error) {
	backends := cluster.BackendsFromContext(req.Context)

	// Create response channel
	results := make(chan dispatchResult, len(backends))

	// Fanout to all backends
	for _, backend := range backends {
		// Set backend for next middlewares
		ctx := cluster.ContextWithBackend(
			req.Context,
			backend)
		backendReq := &cluster.Request{req.Request, ctx}
		go dispatch(results, next, backendReq)
	}
	close(results)

	// Collect and merge responses
	var response *cluster.Response
	for result := range results {
		if result.error != nil {
			return nil, result.error
		}

		if response == nil {
			response = result.response
		} else {
			if err := response.Merge(result.response); err != nil {
				return nil, err
			}
		}
	}

	return response, nil
}

// Call next middleware with requst and push
// response into
func dispatch(
	results chan dispatchResult,
	handler cluster.RequestHandler,
	req *cluster.Request) {
	// Call next middleware
	response, err := handler(req)
	results <- dispatchResult{response, err}
}
