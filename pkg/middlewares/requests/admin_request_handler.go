package requests

import (
	"context"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
)

// AdminHandler will handle all meetings related API requests
type AdminHandler struct {
	router *cluster.Router
}

// AdminRequestHandler creates a new request middleware for handling
// all requests related to meetings.
func AdminRequestHandler(
	router *cluster.Router,
) cluster.RequestMiddleware {
	h := &AdminHandler{
		router: router,
	}
	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return func(
			ctx context.Context,
			req *bbb.Request,
		) (bbb.Response, error) {
			// Dispatch API resources
			switch req.Resource {
			case bbb.ResourceIndex:
				return h.Version(ctx, req)
			case bbb.ResourceGetDefaultConfigXML:
				return h.GetDefaultConfigXML(ctx, req)
			case bbb.ResourceSetConfigXML:
				return h.SetConfigXML(ctx, req)
			}
			// Invoke next middlewares
			return next(ctx, req)
		}
	}
}

// Version responds with the current version. This request
// will not hit a real backend and is not part of the
// API interface.
func (h *AdminHandler) Version(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	res := &bbb.XMLResponse{
		Returncode: "SUCCESS",
		Version:    "2.0",
	}
	res.SetStatus(200)
	return res, nil
}

// GetDefaultConfigXML will lookup a backend for the request
// and will invoke the backend.
func (h *AdminHandler) GetDefaultConfigXML(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	backend, err := h.router.LookupBackend(ctx, req)
	if err != nil {
		return nil, err
	}
	if backend != nil {
		return backend.GetDefaultConfigXML(ctx, req)
	}
	return unknownMeetingResponse(), nil
}

// SetConfigXML will lookup a backend for the request
// and will invoke the backend.
func (h *AdminHandler) SetConfigXML(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	backend, err := h.router.LookupBackend(ctx, req)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return backend.SetConfigXML(ctx, req)
	}
	return unknownMeetingResponse(), nil
}
