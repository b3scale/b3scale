package requests

import (
	"context"
	"net/http"

	sq "github.com/Masterminds/squirrel"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
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

// notImplementedResponse is a placeholder error
func notImplementedResponse() *bbb.XMLResponse {
	res := &bbb.XMLResponse{
		Returncode: bbb.RetFailed,
		Message:    "The api endpoint is not yet implemented",
		MessageKey: "notImplemented",
	}
	res.SetStatus(http.StatusOK) // Prevent auth error...
	return res
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

	playbackBaseURL, hasPlaybackBaseURL := config.GetEnvOpt(
		config.EnvPlaybackBaseURL)

	meetingIDs, hasMeetingIDs := req.Params.MeetingIDs()

	qry := store.QueryRecordingsByFrontendKey(req.Frontend.Key)

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
		rec := state.Recording
		if hasPlaybackBaseURL {
			rec.SetPlaybackBaseURL(playbackBaseURL)
		}

		recordings = append(recordings, state.Recording)
	}

	// Create response with all meetings
	res := &bbb.GetRecordingsResponse{
		XMLResponse: &bbb.XMLResponse{
			Returncode: bbb.RetSuccess,
		},
		Recordings: recordings,
	}
	res.SetStatus(http.StatusOK)
	return res, nil
}

// PublishRecordings will move recordings from the unpublished
// directory into the published will update the state.
func (h *RecordingsHandler) PublishRecordings(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {

	recordIDs, hasRecordIDs := req.Params.RecordIDs()
	if !hasRecordIDs {
		return unknownRecordingResponse(), nil
	}

	conn := store.ConnectionFromContext(ctx)

	for _, id := range recordIDs {
		tx, err := conn.Begin(ctx)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)

		rec, err := store.GetRecordingStateByID(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		if rec == nil {
			return unknownRecordingResponse(), nil
		}

		if err := rec.PublishFiles(); err != nil {
			return nil, err
		}

		// Update state
		rec.Recording.Published = true
		if err := rec.Save(ctx, tx); err != nil {
			return nil, err
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}

	}

	res := &bbb.PublishRecordingsResponse{
		XMLResponse: &bbb.XMLResponse{
			Returncode: bbb.RetSuccess,
		},
		Published: true,
	}
	res.SetStatus(http.StatusOK)

	return res, nil
}

// UpdateRecordings will lookup a backend for the request
// and will invoke the backend.
func (h *RecordingsHandler) UpdateRecordings(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	recordIDs, hasRecordIDs := req.Params.RecordIDs()
	if !hasRecordIDs {
		return unknownRecordingResponse(), nil
	}
	conn := store.ConnectionFromContext(ctx)
	meta := req.Params.ToMetadata()

	for _, recordID := range recordIDs {
		tx, err := conn.Begin(ctx)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)

		rec, err := store.GetRecordingStateByID(ctx, tx, recordID)
		if err != nil {
			return nil, err
		}
		if rec == nil {
			return unknownRecordingResponse(), nil
		}
		// Update metadata
		rec.Recording.Metadata.Update(meta)

		if err := rec.Save(ctx, tx); err != nil {
			return nil, err
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
	}

	// Create response with all meetings
	res := &bbb.UpdateRecordingsResponse{
		XMLResponse: &bbb.XMLResponse{
			Returncode: bbb.RetSuccess,
		},
		Updated: true,
	}
	res.SetStatus(http.StatusOK)
	return res, nil
}

// DeleteRecordings will lookup a backend for the request
// and will invoke the backend.
func (h *RecordingsHandler) DeleteRecordings(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	recordIDs, hasRecordIDs := req.Params.RecordIDs()
	if !hasRecordIDs {
		return unknownRecordingResponse(), nil
	}

	conn := store.ConnectionFromContext(ctx)

	for _, recordID := range recordIDs {
		tx, err := conn.Begin(ctx)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)

		rec, err := store.GetRecordingState(
			ctx, tx, store.QueryRecordingsByFrontendKey(req.Frontend.Key).
				Where("recordings.record_id = ?", recordID))
		if err != nil {
			return nil, err
		}
		if rec == nil {
			return unknownRecordingResponse(), nil
		}

		// Delete recording state.
		if err := store.DeleteRecordingByID(ctx, tx, recordID); err != nil {
			return nil, err
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}

		// Remove from FS
		if err := rec.DeleteFiles(); err != nil {
			return nil, err
		}
	}

	res := &bbb.DeleteRecordingsResponse{
		Deleted: true,
	}
	res.SetStatus(http.StatusOK)

	return res, nil
}

// GetRecordingTextTracks will be passed through to the backend
func (h *RecordingsHandler) GetRecordingTextTracks(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	// recordID, _ := req.Params.RecordID()
	res := notImplementedResponse()
	return res, nil
}

// PutRecordingTextTrack will be passed through to the backend
func (h *RecordingsHandler) PutRecordingTextTrack(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	// recordID, _ := req.Params.RecordID()
	res := notImplementedResponse()
	return res, nil
}
