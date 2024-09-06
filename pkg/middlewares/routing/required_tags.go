package routing

import (
	"context"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
)

// RequiredTags filters backends that match required
// tags defined in the frontend settings by the variables
//
//	required_tags = ["sip", "foo"]
func RequiredTags(next cluster.RouterHandler) cluster.RouterHandler {
	return func(
		ctx context.Context,
		backends []*cluster.Backend,
		req *bbb.Request,
	) ([]*cluster.Backend, error) {

		// This middleware only applies to create meeting requests
		if req.Resource != bbb.ResourceCreate {
			return next(ctx, backends, req) // pass
		}

		// Get tags from frontend settings and filter backends
		frontend := cluster.FrontendFromContext(ctx)
		if frontend == nil {
			return next(ctx, backends, req) // pass
		}

		tags := frontend.Settings().RequiredTags
		backends = filterRequiredTags(backends, tags)

		return next(ctx, backends, req)
	}
}

// filterRequiredTags retrievs the required tags
// for a frontend from the configuration state and
// removes backends not providing all of the tags
func filterRequiredTags(
	backends []*cluster.Backend,
	required []string,
) []*cluster.Backend {
	filtered := make([]*cluster.Backend, 0, len(backends))
	for _, be := range backends {
		if be.HasTags(required) {
			filtered = append(filtered, be)
		}
	}
	return filtered
}
