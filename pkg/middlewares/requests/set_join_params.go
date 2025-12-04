package requests

import (
	"context"
	"strings"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
)

// SetJoinParams produces a middleware setting default
// or overriding parameters in a room join request.
func SetJoinParams() cluster.RequestMiddleware {
	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return func(ctx context.Context, req *bbb.Request) (bbb.Response, error) {
			frontend := cluster.FrontendFromContext(ctx)
			if frontend == nil {
				return next(ctx, req) // pass
			}
			if req.Resource != bbb.ResourceJoin {
				return next(ctx, req) // pass, nothing to do here
			}
			updateJoinParams(req, frontend)
			return next(ctx, req)
		}
	}
}

// updateJoinParams applies parameter overrides and
// adds default values.
func updateJoinParams(req *bbb.Request, fe *cluster.Frontend) {
	defaults := fe.Settings().JoinDefaultParams
	overrides := fe.Settings().JoinOverrideParams

	// Override parameters
	for k, v := range overrides {
		req.Params[k] = v
	}

	// Apply defaults
	for k, v := range defaults {
		if k == bbb.ParamDisabledFeatures {
			continue // special case
		}
		_, ok := req.Params[k]
		if !ok {
			req.Params[k] = v // set if not present
		}
	}

	// Handle special parameters
	updateJoinDisabledFeatures(req, fe)
}

// updateJoinDisabledFeatures updates the disabledFeatures
// parameter of a request if a default is present.
func updateJoinDisabledFeatures(req *bbb.Request, fe *cluster.Frontend) {
	defaults := fe.Settings().JoinDefaultParams
	disabledFeaturesDefaultParam, ok := defaults[bbb.ParamDisabledFeatures]
	if !ok {
		return // Nothing to update
	}
	disabledFeaturesRequestParam, ok := req.Params[bbb.ParamDisabledFeatures]
	if !ok {
		disabledFeaturesRequestParam = ""
	}

	disabledDefault := strings.Split(disabledFeaturesDefaultParam, ",")
	disabledRequest := strings.Split(disabledFeaturesRequestParam, ",")

	// Merge, deduplicate and filter empty
	disabledRequest = append(disabledRequest, disabledDefault...)
	disabledFeatures := make([]string, 0, len(disabledRequest))
	found := map[string]bool{}
	for _, feature := range disabledRequest {
		if feature == "" {
			continue
		}
		if _, ok := found[feature]; ok {
			continue
		}
		disabledFeatures = append(disabledFeatures, feature)
		found[feature] = true
	}
	req.Params[bbb.ParamDisabledFeatures] = strings.Join(disabledFeatures, ",")
}
