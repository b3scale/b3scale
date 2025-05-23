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
			if err := rewriteCallbacks(ctx, req); err != nil {
				return nil, err
			}
			return next(ctx, req)
		}
	}
}

// Rewrite all well known callback parameters
func rewriteCallbacks(
	ctx context.Context,
	req *bbb.Request,
) error {
	callbacks := []string{
		bbb.MetaParamRecordingReadyURL,
		bbb.MetaParamAnalyticsCallbackURL,
		bbb.MetaParamMeetingEndCallbackURL,
		bbb.ParamMeetingEndedURL,
	}
	for _, callback := range callbacks {
		if err := rewriteCallback(ctx, req, callback); err != nil {
			return err
		}
	}
	return nil
}

// Rewrite meta callback URLs: Callback URLs are rewritten
// to point to a b3scale endpoin. The original URL and callback
// is encoded as the AUD attribute of the token.
func rewriteCallback(
	ctx context.Context,
	req *bbb.Request,
	callback string,
) error {
	apiURL := config.MustEnv(config.EnvAPIURL)
	secret := config.MustEnv(config.EnvJWTSecret)

	// Check if we have a known callback URL
	callbackURL, ok := req.Params[callback]
	if !ok {
		return nil // nothing to do here
	}

	frontend := cluster.FrontendFromContext(ctx)
	if frontend == nil {
		return nil
	}

	// Encode URL in token
	token, err := auth.NewClaims(frontend.ID()).
		WithAudience(callbackURL).
		WithScopes(auth.ScopeCallback).
		Sign(secret)
	if err != nil {
		return err
	}

	// Rewrite to our own endpoint
	proxyURL := fmt.Sprintf(
		"%s/api/v1/callbacks/proxy/%s",
		apiURL,
		token)

	req.Params[callback] = proxyURL
	return nil
}
