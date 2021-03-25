package requests

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// DefaultPresentation produces a middleware for injecting
// a XML snippet into the request body of a create request.
// There are two frontend setting variables:
//
//   default_presentation.url = https://path-to-presentation
//   default_presentation.force = true | false
//
func DefaultPresentation() cluster.RequestMiddleware {
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
	// Get settings
	var (
		presentationURL string = fe.Settings().GetString("default_presentation.url", "")
		force           string = fe.Settings().GetString("default_presentation.force", "")
	)

	// Are we active?
	if presentationURL == "" {
		return // nothing to do here
	}

	// Is this a create request?
	if req.Resource != bbb.ResourceCreate {
		return // Nothing to do here
	}

	// We have a presentation URL, let's check if there is already
	// a request body present
	if req.HasBody() && force != "true" {
		return // presentation present, nothing to do here
	}

	// Set or override presentation
	req.Header.Set("content-type", "application/xml")
	req.Body = makePresentationRequestBody(presentationURL)
}

// create the XML document
func makePresentationRequestBody(url string) []byte {
	tmpl := strings.TrimSpace(`
<modules>
   <module name="presentation">
      <document url="%s" filename="%s"/>
   </module>
</modules>`)

	filename := filepath.Base(url)
	body := fmt.Sprintf(tmpl, url, filename)
	return []byte(body)
}
