package requests

import (
	"context"
	"net/http"

	sq "github.com/Masterminds/squirrel"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
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

// unknownRecordingResponse is a standard error response,
// when a recording could not be found by a lookup.
func unknownRecordingResponse() *bbb.XMLResponse {
	res := &bbb.XMLResponse{
		Returncode: bbb.RetFailed,
		Message:    "The recording is not known to us.",
		MessageKey: "invalidRecordingIdentifier",
	}
	res.SetStatus(http.StatusOK) // I'm pretty sure we need
	// to respond with some success status code, otherwise
	// greenlight and the like will assume incorrect credentials
	// or something.
	return res
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

// GetRecordings will retrieve all recordings for
// the given frontend instance.
func (h *RecordingsHandler) GetRecordings(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	meetingIDs, hasMeetingIDs := req.Params.MeetingIDs()

	qry := store.Q().
		Join("frontends ON frontends.id = recordings.frontend_id").
		Where("recordings.frontend_id IS NOT NULL").
		Where("frontends.key = ?", req.Frontend.Key)

	if hasMeetingIDs {
		filterMIDs := sq.Or{}
		for _, mid := range meetingIDs {
			filterMIDs = append(filterMIDs, sq.Eq{
				"recordings.meeting_id": mid,
			})
		}
		qry = qry.Where(filterMIDs)
	}

	recordingStates, err := store.GetRecordingStates(ctx, tx, qry)
	if err != nil {
		return nil, err
	}
	tx.Rollback(ctx)

	recordings := make([]*bbb.Recording, 0, len(recordingStates))
	for _, state := range recordingStates {
		recordings = append(recordings, state.Recording)
	}

	// Create response with all meetings
	res := &bbb.GetRecordingsResponse{
		XMLResponse: &bbb.XMLResponse{
			Returncode: "SUCCESS",
		},
		Recordings: recordings,
	}
	res.SetStatus(http.StatusOK)
	return res, nil
}

// PublishRecordings will lookup a backend for the request
// and will invoke the backend.
func (h *RecordingsHandler) PublishRecordings(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	var beRes bbb.Response

	recordIDs, hasRecordIDs := req.Params.RecordIDs()
	if !hasRecordIDs {
		return unknownRecordingResponse(), nil
	}

	for _, recordID := range recordIDs {
		backend, err := h.router.LookupBackendForRecordID(ctx, recordID)
		if err != nil {
			return nil, err
		}

		beReq := bbb.PublishRecordingRequest(recordID, req.Params)
		res, err := backend.PublishRecordings(ctx, beReq)
		if err != nil {
			return nil, err
		}
		if !res.IsSuccess() {
			return res, nil
		}

		err = backend.RefreshRecording(ctx, recordID)
		if err != nil {
			return nil, err
		}

		beRes = res
	}
	return beRes, nil
}

// UpdateRecordings will lookup a backend for the request
// and will invoke the backend.
func (h *RecordingsHandler) UpdateRecordings(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	var beRes bbb.Response

	recordIDs, hasRecordIDs := req.Params.RecordIDs()
	if !hasRecordIDs {
		return unknownRecordingResponse(), nil
	}

	for _, recordID := range recordIDs {
		backend, err := h.router.LookupBackendForRecordID(ctx, recordID)
		if err != nil {
			return nil, err
		}

		beReq := bbb.UpdateRecordingRequest(recordID, req.Params)
		res, err := backend.UpdateRecordings(ctx, beReq)
		if err != nil {
			return nil, err
		}
		if !res.IsSuccess() {
			return res, nil
		}

		err = backend.RefreshRecording(ctx, recordID)
		if err != nil {
			return nil, err
		}

		beRes = res
	}
	return beRes, nil
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
	recordID, ok := req.Params.RecordID()
	if !ok {
		return unknownRecordingResponse(), nil
	}

	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	tracks, err := store.GetRecordingTextTracks(ctx, tx, recordID)
	if err != nil {
		return nil, err
	}

	res := &bbb.GetRecordingTextTracksResponse{
		Returncode: bbb.RetSuccess,
		Tracks:     tracks,
	}
	res.SetStatus(http.StatusOK)

	return res, nil
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
