package requests

import (
	"context"
	"net/http"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// RecordingsHandlerOptions has configuration options for
// this middleware handling all recordings.
type RecordingsHandlerOptions struct {
	// Most likely some NFS shared storage or s3
	// config or whatever...
}

// RecordingsHandler will handle all meetings related API requests
type RecordingsHandler struct {
	opts   *RecordingsHandlerOptions
	router *cluster.Router
}

// RecordingsRequestHandler creates a new request middleware for handling
// all requests related to meetings.
func RecordingsRequestHandler(
	router *cluster.Router,
	opts *RecordingsHandlerOptions,
) cluster.RequestMiddleware {
	h := &RecordingsHandler{
		opts:   opts,
		router: router,
	}
	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return func(
			ctx context.Context,
			req *bbb.Request,
		) (bbb.Response, error) {
			switch req.Resource {
			case bbb.ResourceGetRecordings:
				return h.GetRecordings(ctx, req)
			case bbb.ResourcePublishRecordings:
				return h.PublishRecordings(ctx, req)
			case bbb.ResourceDeleteRecordings:
				return h.DeleteRecordings(ctx, req)
			case bbb.ResourceUpdateRecordings:
				return h.UpdateRecordings(ctx, req)
			case bbb.ResourceGetRecordingTextTracks:
				return h.GetRecordingTextTracks(ctx, req)
			case bbb.ResourcePutRecordingTextTrack:
				return h.PutRecordingTextTrack(ctx, req)
			}
			// Invoke next middlewares
			return next(ctx, req)
		}
	}
}

// GetRecordings will lookup a backend for the request
// and will invoke the backend.
func (h *RecordingsHandler) GetRecordings(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	noRecordingsRes := &bbb.GetRecordingsResponse{
		XMLResponse: &bbb.XMLResponse{
			Returncode: bbb.RetSuccess,
		},
	}
	noRecordingsRes.SetStatus(http.StatusOK)

	backend, err := h.router.LookupBackend(ctx, req)
	if err != nil {
		return nil, err
	}
	if backend == nil {
		return noRecordingsRes, nil
	}

	res, err := backend.GetRecordings(ctx, req)
	if err != nil {
		// Return failed successfully response
		return noRecordingsRes, nil
	}

	return res, nil
}

// PublishRecordings will lookup a backend for the request
// and will invoke the backend.
func (h *RecordingsHandler) PublishRecordings(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	backend, err := h.router.LookupBackend(ctx, req)
	if err != nil {
		return nil, err
	}
	if backend != nil {
		return backend.PublishRecordings(ctx, req)
	}
	return unknownMeetingResponse(), nil
}

// UpdateRecordings will lookup a backend for the request
// and will invoke the backend.
func (h *RecordingsHandler) UpdateRecordings(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	backend, err := h.router.LookupBackend(ctx, req)
	if err != nil {
		return nil, err
	}
	if backend != nil {
		return backend.UpdateRecordings(ctx, req)
	}
	return unknownMeetingResponse(), nil
}

// DeleteRecordings will lookup a backend for the request
// and will invoke the backend.
func (h *RecordingsHandler) DeleteRecordings(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	backend, err := h.router.LookupBackend(ctx, req)
	if err != nil {
		return nil, err
	}
	if backend != nil {
		return backend.GetRecordings(ctx, req)
	}
	return unknownMeetingResponse(), nil
}

// GetRecordingTextTracks will lookup a backend for the request
// and will invoke the backend.
func (h *RecordingsHandler) GetRecordingTextTracks(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	backend, err := h.router.LookupBackend(ctx, req)
	if err != nil {
		return nil, err
	}
	if backend != nil {
		return backend.GetRecordingTextTracks(ctx, req)
	}
	return unknownMeetingResponse(), nil
}

// PutRecordingTextTrack will lookup a backend for the request
// and will invoke the backend.
func (h *RecordingsHandler) PutRecordingTextTrack(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	backend, err := h.router.LookupBackend(ctx, req)
	if err != nil {
		return nil, err
	}
	if backend != nil {
		return backend.PutRecordingTextTrack(ctx, req)
	}
	return unknownMeetingResponse(), nil
}
