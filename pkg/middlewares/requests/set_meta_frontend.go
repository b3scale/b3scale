package requests

import (
	"context"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// SetMetaFrontend creates a middleware for adding
// a `meta_frontend` parameter to the create request.
func SetMetaFrontend() cluster.RequestMiddleware {
	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return func(ctx context.Context, req *bbb.Request) (bbb.Response, error) {
			maybeSetMetaFrontend(ctx, req)
			return next(ctx, req)
		}
	}
}

// In case of a create request, an additional meta
// parameter `meta_frontend` will be injected.
func maybeSetMetaFrontend(ctx context.Context, req *bbb.Request) {
	frontend := cluster.FrontendFromContext(ctx)
	if frontend == nil {
		return
	}
	if req.Resource != bbb.ResourceCreate {
		return // Nothing to do here.
	}
	req.Params[bbb.MetaParam("frontend")] = frontend.Key()
}
