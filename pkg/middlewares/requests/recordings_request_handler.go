package requests

import (
	"context"
	"fmt"
	"net/http"

	sq "github.com/Masterminds/squirrel"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/store"
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

// GetRecordings filters: filter by state
func maybeFilterRecordingStates(
	qry sq.SelectBuilder,
	params bbb.Params,
) sq.SelectBuilder {
	states, ok := params.States()
	if !ok {
		return qry
	}

	filters := sq.Or{}
	for _, s := range states {
		if s == bbb.StateAny {
			return qry // nothing to filter
		}
		filters = append(filters, sq.Eq{
			"recordings.state -> 'state'": s,
		})
	}

	return qry.Where(filters)
}

// GetRecordings filters: filter by meeting ids
func maybeFilterRecordingMeetingIDs(
	qry sq.SelectBuilder,
	params bbb.Params,
) sq.SelectBuilder {
	meetingIDs, ok := params.MeetingIDs()
	if !ok {
		return qry // nothing to do here
	}
	filters := sq.Or{}
	for _, mid := range meetingIDs {
		filters = append(filters, sq.Eq{
			"recordings.meeting_id": mid,
		})
	}
	return qry.Where(filters)
}

// GetRecordings filters: filter by set of recordID.
// The recordID can also be used as a wildcard by including
// only the first characters in the string. I interpret this
// as a 'LIKE' query...
func maybeFilterRecordingIDs(
	qry sq.SelectBuilder,
	params bbb.Params,
) sq.SelectBuilder {
	recordIDs, ok := params.RecordIDs()
	if !ok {
		return qry
	}
	filters := sq.Or{}
	for _, rid := range recordIDs {
		filters = append(filters, sq.Like{
			"recordings.record_id": rid + "%",
		})
	}
	return qry.Where(filters)
}

// GetRecordings filters: filter by metadata
func maybeFilterRecordingMeta(
	qry sq.SelectBuilder,
	params bbb.Params,
) sq.SelectBuilder {
	meta := params.ToMetadata()
	if len(meta) == 0 {
		return qry
	}
	filters := sq.And{}
	for k, v := range meta {
		filters = append(filters, sq.Eq{
			fmt.Sprintf(
				"recordings.state -> 'metadata' -> '%s'",
				store.SQLSafeParam(k)): v,
		})
	}
	return qry.Where(filters)
}

// GetRecordings will retrieve all recordings for
// the given frontend instance. The use of 'state' might be
// a bit confusing here:
// In the database, 'state' refers to the attribute holding the
// acutual recording JSON object.
// A recording itself can also have a state attribute
// (like published, unpublished).
func (h *RecordingsHandler) GetRecordings(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	playbackHost, hasPlaybackHost := config.GetEnvOpt(
		config.EnvRecordingsPlaybackHost)

	qry := store.QueryRecordingsByFrontendKey(req.Frontend.Key)

	// Apply filters to query: The API supports search by
	// meetingIDs, states, recordIDs and metadata.
	qry = maybeFilterRecordingIDs(qry, req.Params)
	qry = maybeFilterRecordingMeetingIDs(qry, req.Params)
	qry = maybeFilterRecordingStates(qry, req.Params)
	qry = maybeFilterRecordingMeta(qry, req.Params)

	recordingStates, err := store.GetRecordingStates(ctx, tx, qry)
	if err != nil {
		return nil, err
	}

	recordings := make([]*bbb.Recording, 0, len(recordingStates))
	for _, state := range recordingStates {
		rec := state.Recording
		if hasPlaybackHost {
			rec.SetPlaybackHost(playbackHost)
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
	conn := store.ConnectionFromContext(ctx)

	recordIDs, hasRecordIDs := req.Params.RecordIDs()
	if !hasRecordIDs {
		return unknownRecordingResponse(), nil
	}
	publish, _ := req.Params.Publish()

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

		if publish {
			if err := rec.PublishFiles(); err != nil {
				return nil, err
			}
			rec.Recording.State = bbb.StatePublished
			rec.Recording.Published = true
		} else {
			if err := rec.UnpublishFiles(); err != nil {
				return nil, err
			}
			rec.Recording.State = bbb.StateUnpublished
			rec.Recording.Published = false
		}

		// Persist changes
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
		Published: publish,
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
	conn := store.ConnectionFromContext(ctx)

	recordIDs, hasRecordIDs := req.Params.RecordIDs()
	if !hasRecordIDs {
		return unknownRecordingResponse(), nil
	}

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

		// Remove from FS (this will fail more likely than deleting
		// the database record, so we do this first in case this fails.
		if err := rec.DeleteFiles(); err != nil {
			return nil, err
		}

		// Delete recording state.
		if err := store.DeleteRecordingByID(ctx, tx, recordID); err != nil {
			return nil, err
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
	}

	res := &bbb.DeleteRecordingsResponse{
		XMLResponse: &bbb.XMLResponse{
			Returncode: bbb.RetSuccess,
		},
		Deleted: true,
	}
	res.SetStatus(http.StatusOK)
	return res, nil
}

// GetRecordingTextTracks seems to be not implemented
// currently in scalelite. We will try to figure out later
// what is actually supposed to be going on here.
func (h *RecordingsHandler) GetRecordingTextTracks(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	res := notImplementedResponse()
	return res, nil
}

// PutRecordingTextTrack will also be implemented later.
func (h *RecordingsHandler) PutRecordingTextTrack(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	res := notImplementedResponse()
	return res, nil
}
