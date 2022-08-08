package requests

import (
	"context"
	"strings"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
)

// SetCreateParams produces a middleware setting default
// or overriding parameters in a room create request.
func SetCreateParams() cluster.RequestMiddleware {
	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return func(ctx context.Context, req *bbb.Request) (bbb.Response, error) {
			frontend := cluster.FrontendFromContext(ctx)
			if frontend == nil {
				return next(ctx, req) // pass
			}
			if req.Resource != bbb.ResourceCreate {
				return next(ctx, req) // pass, nothing to do here
			}
			updateCreateParams(req, frontend)
			return next(ctx, req)
		}
	}
}

// updateCreateParams applies parameter overrides and
// adds default values.
func updateCreateParams(req *bbb.Request, fe *cluster.Frontend) {
	defaults := fe.Settings().CreateDefaultParams
	overrides := fe.Settings().CreateOverrideParams

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
	updateCreateDisabledFeatures(req, fe)
}

// updateCreateDisabledFeatures updates the disabledFeatures
// paramter of a request if a default is present.
func updateCreateDisabledFeatures(req *bbb.Request, fe *cluster.Frontend) {
	defaults := fe.Settings().CreateDefaultParams
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
