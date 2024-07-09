package requests

import (
	"context"
	"fmt"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/auth"
)

// RewriteMetaCallbackURLs creates a middleware for rewriting
// meta parameters like `meta_bbb-recording-ready-url` to a
// local endpoint.
func RewriteMetaCallbackURLs() cluster.RequestMiddleware {
	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return func(ctx context.Context, req *bbb.Request) (bbb.Response, error) {
			if err := rewriteRecordingReadyURL(ctx, req); err != nil {
				return nil, err
			}
			return next(ctx, req)
		}
	}
}

// Rewrite meta parameters when present.
func rewriteRecordingReadyURL(ctx context.Context, req *bbb.Request) error {
	apiURL := config.MustEnv(config.EnvAPIURL)
	secret := config.MustEnv(config.EnvJWTSecret)

	readyURL, ok := req.Params[bbb.MetaParamRecordingReadyURL]
	if !ok {
		return nil // nothing to do here
	}

	frontend := cluster.FrontendFromContext(ctx)
	if frontend == nil {
		return nil
	}

	// Encode URL in token
	token, err := auth.NewClaims(frontend.ID()).
		WithAudience(readyURL).
		WithScopes(auth.ScopeCallback).
		Sign(secret)
	if err != nil {
		return err
	}

	// Rewrite to our own endpoint
	callbackURL := fmt.Sprintf(
		"%s/api/v1/recordings/ready/%s",
		apiURL,
		token)
	req.Params[bbb.MetaParamRecordingReadyURL] = callbackURL

	return nil
}
