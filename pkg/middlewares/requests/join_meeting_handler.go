package requests

// Join Meetings Handler Middleware
// --------------------------------
//
// This is a handler for 'join meeting' requests.
//
// If the meeting is not yet available we stall.
// If the backend is not available we stall.
// We do this by redirecting to a waiting page and
// reissue the request after some seconds.
//

import (
	"context"
	"net/http"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
	"gitlab.com/infra.run/public/b3scale/pkg/templates"
)

// JoinMeetingHandlerOptions has configuration options for
// this middleware.
type JoinMeetingHandlerOptions struct {
	// UseReverseProxy will handle the request in a way
	// that a reverse proxy can be used. This is an experimental
	// feature and a known issue is the unfortunate handling
	// of breakout rooms.
	UseReverseProxy bool
}

// JoinMeetingHandler creates a new request middleware for
// handling join requests.
func JoinMeetingHandler(opts *JoinMeetingHandlerOptions) cluster.RequestMiddleware {
	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return func(
			ctx context.Context,
			req *bbb.Request,
		) (bbb.Response, error) {
			if req.Resource == bbb.ResourceJoin {
				return handleJoinMeeting(ctx, req, opts)
			}
			// Invoke next middlewares. (This would be pressumably
			// normal routing and dispatching.)
			return next(ctx, req)
		}
	}
}

// handleJoinMeeting will try to join the meeting
// provided by meetingID - or stall in case the meeting is not yet
// or the backend is temporary not available.
func handleJoinMeeting(
	ctx context.Context,
	req *bbb.Request,
	opts *JoinMeetingHandlerOptions,
) (bbb.Response, error) {
	frontend := cluster.FrontendFromContext(ctx)

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

	// Assign frontend if not present. As we disallow rebinding
	// this will fail if the frontendID not null or the same
	// as the request frontendID.
	if err := meeting.BindFrontendID(ctx, tx, frontend.ID()); err != nil {
		return nil, err
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

	// Dispatch to backend
	backend := cluster.NewBackend(backendState)
	if opts.UseReverseProxy {
		return backend.JoinProxy(ctx, req)
	}
	return backend.Join(ctx, req)
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
