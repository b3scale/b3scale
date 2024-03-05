package requests

import (
	"context"
	"path/filepath"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/templates"
)

// SetDefaultPresentation produces a middleware for injecting
// a XML snippet into the request body of a create request.
// There are two frontend setting variables:
//
//	default_presentation.url = https://path-to-presentation
//	default_presentation.force = true | false
func SetDefaultPresentation() cluster.RequestMiddleware {
	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return func(ctx context.Context, req *bbb.Request) (bbb.Response, error) {
			frontend := cluster.FrontendFromContext(ctx)
			if frontend == nil {
				return next(ctx, req) // pass
			}
			maybeUpdateDefaultPresentation(req, frontend)
			return next(ctx, req)
		}
	}
}

func maybeUpdateDefaultPresentation(req *bbb.Request, fe *cluster.Frontend) {
	opts := fe.Settings().DefaultPresentation

	// Are we active?
	if opts == nil || opts.URL == "" {
		return // nothing to do here
	}

	// Is this a create request?
	if req.Resource != bbb.ResourceCreate {
		return // Nothing to do here
	}

	// We have a presentation URL, let's check if there is already
	// a request body present
	if req.HasBody() && !opts.Force {
		return // presentation present, nothing to do here
	}

	presURL := opts.URL
	filename := filepath.Base(presURL)

	// Set or override presentation
	req.Header.Set("content-type", "application/xml")
	req.Body = templates.DefaultPresentationBody(presURL, filename)
}
