package requests

import (
	"context"
	"errors"
	"net/http"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
	"gitlab.com/infra.run/public/b3scale/pkg/templates"
)

// MeetingsHandlerOptions has configuration options for
// this middleware handling all meetings related stuff.
type MeetingsHandlerOptions struct {
	// UseReverseProxy will handle the request in a way
	// that a reverse proxy can be used. This is an experimental
	// feature and a known issue is the unfortunate handling of breakout rooms.
	// When deployed in reverse proxy mode we will handle the
	// join internally and the proxy needs to handle subsequent requests.
	UseReverseProxy bool
}

// MeetingsHandler will handle all meetings related API requests
type MeetingsHandler struct {
	opts   *MeetingsHandlerOptions
	router *cluster.Router
}

// MeetingsRequestHandler creates a new request middleware for handling
// all requests related to meetings.
func MeetingsRequestHandler(
	router *cluster.Router,
	opts *MeetingsHandlerOptions,
) cluster.RequestMiddleware {
	h := &MeetingsHandler{
		opts:   opts,
		router: router,
	}

	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return func(
			ctx context.Context,
			req *bbb.Request,
		) (bbb.Response, error) {
			switch req.Resource {
			case bbb.ResourceJoin:
				return h.Join(ctx, req)
			case bbb.ResourceCreate:
				return h.Create(ctx, req)
			case bbb.ResourceIsMeetingRunning:
				return h.IsMeetingRunning(ctx, req)
			case bbb.ResourceEnd:
				return h.End(ctx, req)
			case bbb.ResourceGetMeetingInfo:
				return h.GetMeetingInfo(ctx, req)
			case bbb.ResourceGetMeetings:
				return h.GetMeetings(ctx, req)
			}
			// Invoke next middlewares
			return next(ctx, req)
		}
	}
}

// Join will try to join the meeting
//
// If the meeting is not yet available we stall. If the backend is not
// available we stall.
// We do this by redirecting to a waiting page and reissue the request
// after some seconds.
func (h *MeetingsHandler) Join(
	ctx context.Context, req *bbb.Request,
) (bbb.Response, error) {
	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Lookup meeting
	meetingID, _ := req.Params.MeetingID()
	meeting, err := store.GetMeetingStateByID(ctx, tx, meetingID)
	if err != nil {
		return nil, err
	}
	if meeting == nil {
		// The meeting is now known to the cluster.
		// Send user to the stalling page.
		return retryJoinResponse(req), nil
	}

	// Get backend do redirect
	backendState, err := meeting.GetBackendState(ctx, tx)
	if err != nil {
		return nil, err
	}

	// In case the meeting is not assigned to backend (yet)
	if backendState == nil {
		return retryJoinResponse(req), nil
	}

	// We have a backend - yay! check that the backend is
	// ok and the node agent is alive.
	if !backendState.IsNodeReady() {
		return retryJoinResponse(req), nil
	}

	// Commit changes
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Dispatch to backend
	backend := cluster.NewBackend(backendState)
	if h.opts.UseReverseProxy {
		return backend.JoinProxy(ctx, req)
	}
	return backend.Join(ctx, req)
}

// Create will acquire a backend from the router
// selected for the request and create the meeting.
func (h *MeetingsHandler) Create(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	var (
		backend *cluster.Backend
		err     error
	)
	// Lookup backend, as we need to make this
	// endpoint idempotent
	backend, err = h.router.LookupBackend(ctx, req)
	if errors.Is(err, cluster.ErrNoBackendForMeeting) {
		backend, err = h.router.SelectBackend(ctx, req)
	}
	if err != nil {
		return nil, err
	}
	return backend.Create(ctx, req)
}

// IsMeetingRunning will check on a backend if the meeting is still running
func (h *MeetingsHandler) IsMeetingRunning(
	ctx context.Context, req *bbb.Request,
) (bbb.Response, error) {
	backend, err := h.router.LookupBackend(ctx, req)
	if errors.Is(err, cluster.ErrNoBackendForMeeting) {
		// Return failed successfully response
		res := &bbb.IsMeetingRunningResponse{
			XMLResponse: &bbb.XMLResponse{
				Returncode: bbb.RetSuccess,
			},
			Running: false,
		}
		res.SetStatus(http.StatusOK)
		return res, nil
	}
	if err != nil {
		return nil, err
	}
	return backend.IsMeetingRunning(ctx, req)
}

// End will end a meeting on a backend
func (h *MeetingsHandler) End(
	ctx context.Context, req *bbb.Request,
) (bbb.Response, error) {
	backend, err := h.router.LookupBackend(ctx, req)
	if err != nil {
		return nil, err
	}
	return backend.End(ctx, req)
}

// GetMeetingInfo will not hit a backend, but we will query
// the store directly.
func (h *MeetingsHandler) GetMeetingInfo(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	backend, err := h.router.LookupBackend(ctx, req)
	if err != nil {
		return nil, err
	}
	res, err := backend.GetMeetingInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetMeetings lists all meetings in the cluster relevant
// for the frontend
func (h *MeetingsHandler) GetMeetings(
	ctx context.Context,
	req *bbb.Request,
) (bbb.Response, error) {
	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get all meetings from our store associated
	// with the requesting frontend.
	mstates, err := store.GetMeetingStates(ctx, tx, store.Q().
		Join("frontends ON frontends.id = meetings.frontend_id").
		Where("meetings.backend_id IS NOT NULL").
		Where("frontends.key = ?", req.Frontend.Key))
	if err != nil {
		return nil, err
	}

	tx.Rollback(ctx)

	meetings := make([]*bbb.Meeting, 0, len(mstates))
	for _, state := range mstates {
		meetings = append(meetings, state.Meeting)
	}

	// Create response with all meetings
	res := &bbb.GetMeetingsResponse{
		XMLResponse: &bbb.XMLResponse{
			Returncode: "SUCCESS",
		},
		Meetings: meetings,
	}
	res.SetStatus(200)

	return res, nil
}

// retryJoinResponse makes a new JoinResponse with
// a redirect to a waiting page. The original request will be
// encoded and passed to the page as a parameter.
func retryJoinResponse(req *bbb.Request) *bbb.JoinResponse {
	retryURL := "/_b3scale/retry-join/" + string(req.MarshalURLSafe())
	body := templates.Redirect(retryURL)

	// Create custom join response
	res := &bbb.JoinResponse{
		XMLResponse: new(bbb.XMLResponse),
	}
	res.SetStatus(http.StatusFound)
	res.SetRaw(body)
	res.SetHeader(http.Header{
		"content-type": []string{"text/html"},
		"location":     []string{req.URL()},
	})
	return res
}
