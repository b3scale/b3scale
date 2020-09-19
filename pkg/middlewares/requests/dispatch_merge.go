package requests

// Dispatch / Merge Middleware
// Forwards the request to the selected backends
// and merges the response.

import (
	"context"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

type dispatchResult struct {
	response bbb.Response
	error    error
}

// NewDispatchMerge creates the request middleware
// function, dispatching the request to nodes and
// collects responses.
func NewDispatchMerge() cluster.RequestMiddleware {
	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return func(ctx context.Context, req *bbb.Request) (bbb.Response, error) {

			return dispatchMerge(ctx, next, req)
		}
	}
}

// Actual dispatch and merge of the request.
func dispatchMerge(
	ctx context.Context,
	next cluster.RequestHandler,
	req *bbb.Request,
) (bbb.Response, error) {
	// Create response channel for all backends
	backends := cluster.BackendsFromContext(ctx)
	results := make(chan dispatchResult, len(backends))

	// Fanout to all backends
	for _, backend := range backends {
		// Set backend for next middlewares
		ctx := cluster.ContextWithBackend(ctx, backend)
		go dispatch(ctx, results, next, req)
	}
	close(results)

	// Collect and merge responses
	var response bbb.Response
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
	ctx context.Context,
	results chan dispatchResult,
	handler cluster.RequestHandler,
	req *bbb.Request) {
	// Call next middleware
	response, err := handler(ctx, req)
	results <- dispatchResult{response, err}
}
