package requests

import (
	"context"

	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// BindMeetingFrontend asserts, that if a meeting exists
// and the meeting is either not bound to a frontend
// the meetings frontend will be set to the requesting
// frontend.
// We do not support rebinding between frontends and will
// fail if a frontend tries to steal a meeting.
func BindMeetingFrontend() cluster.RequestMiddleware {
	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return func(
			ctx context.Context,
			req *bbb.Request,
		) (bbb.Response, error) {
			meetingID, ok := req.Params.MeetingID()
			if !ok {
				return next(ctx, req) // nothing to do here
			}
			frontend := cluster.FrontendFromContext(ctx)
			if frontend == nil {
				return nil, cluster.ErrNoFrontendInContext
			}

			if err := bindMeetingFrontend(ctx, meetingID, frontend); err != nil {
				return nil, err
			}
			return next(ctx, req)
		}
	}
}

// Bind meeting to frontend. This will fail if the meeting
// is already associated with another frontend.
func bindMeetingFrontend(
	ctx context.Context,
	meetingID string,
	frontend *cluster.Frontend,
) error {
	// Next, we have to access our shared state
	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	meeting, err := store.GetMeetingStateByID(ctx, tx, meetingID)
	if err != nil {
		return err
	}
	if meeting == nil {
		// Unknown meeting. Nothing to really do here.
		return nil
	}
	// Assign frontend if not present
	if err := meeting.BindFrontendID(ctx, tx, frontend.ID()); err != nil {
		return err
	}
	// Persist changes and close transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}
