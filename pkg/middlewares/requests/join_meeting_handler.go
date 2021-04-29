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
	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// JoinMeetingOptions has configuration options for
// this middleware.
type JoinMeetingOptions struct {
	// UseReverseProxy will handle the request in a way
	// that a reverse proxy can be used. This is an experimental
	// feature and a known issue is the unfortunate handling
	// of breakout rooms.
	UseReverseProxy bool
}

// JoinMeetingHandler creates a new request middleware for
// handling join requests.
func JoinMeetingHandler(opts *JoinMeetingOptions) cluster.RequestMiddleware {
	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return func(
			ctx context.Context,
			req *bbb.Request,
		) (bbb.Response, error) {
			if req.Resource == bbb.ResourceJoin {
				return handleJoinMeeting(opts, ctx, req)
			}

			// Invoke next middlewares (this would be pressumably
			// normal routing and dispatching)
			return next(ctx, req)
		}
	}
}

// handleJoinMeeting will try to join the meeting
// provided by meetingID - or stall in case the meeting is
//
