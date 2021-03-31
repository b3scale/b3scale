package routing

import (
	"context"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// RequireTags filters backends that match required
// tags defined in the frontend settings by the variables
//
//   require_tags = ["sip", "foo"]
//
func RequireTags(next cluster.RouterHandler) cluster.RouterHandler {
	return func(
		ctx context.Context,
		backends []*cluster.Backend,
		req *bbb.Request,
	) ([]*cluster.Backend, error) {
		frontend := cluster.FrontendFromContext(ctx)
		if frontend == nil {
			return next(ctx, backends, req) // pass
		}

		return next(ctx, backends, req)
	}
}
