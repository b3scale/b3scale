package requests

import (
	"context"
  "net/http"

	"github.com/jackc/pgx/v4/"
	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/store"
	"github.com/b3scale/b3scale/pkg/templates"
)

// CheckAttendeesLimit produces a middleware for checking
// wether the limit of overall attendees is reached for a frontend.
// There is one frontend setting variables:
//
//	limit_attendees.limit = 10
func CheckAttendeesLimit() cluster.RequestMiddleware {
	return func(next cluster.RequestHandler) cluster.RequestHandler {
		return func(ctx context.Context, req *bbb.Request) (bbb.Response, error) {
			frontend := cluster.FrontendFromContext(ctx)
			if frontend == nil {
				return next(ctx, req) // pass
			}

			// Get database connection
			tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
			if err != nil {
				return nil, err
			}
			defer tx.Rollback(ctx)

			allowed := maybeCheckAttendeesLimit(req, frontend, ctx, tx)
			if !allowed {
				body := templates.AttendeesLimitReached()
				res := &bbb.JoinResponse{
					XMLResponse: &bbb.XMLResponse{
						Returncode: bbb.RetFailed,
						Message:    "The maximum number of participants allowed for this frontend has been reached.",
						MessageKey: "attendeesLimitReached",
					},
				}
				res.SetRaw(body)
				res.SetStatus(403)
				res.SetHeader(http.Header{
					"content-type": []string{"text/html"},
				})
				return res, nil
			}
			return next(ctx, req)
		}
	}
}

func maybeCheckAttendeesLimit(req *bbb.Request, fe *cluster.Frontend, ctx context.Context, tx pgx.Tx) bool {
	opts := fe.Settings().LimitAttendees

	// Are we active?
	if opts == nil || opts.Limit == 0 {
		return true // nothing to do here
	}

	// Is this a join request?
	if req.Resource != bbb.ResourceJoin {
		return true // Nothing to do here
	}

	// Get the sum of the participantCount for all meetings of this meetings frontend
	var curAt int
	qry := `
  select sum((state->>'ParticipantCount')::int) from meetings where frontend_id = $1
  `
	tx.QueryRow(ctx, qry, fe.ID()).Scan(&curAt)

	// If limit was already reached stop request
	if curAt >= opts.Limit {
		log.Info().Str("frontend_key", fe.Key()).Int("limit", opts.Limit).Int("current_attendees", curAt).Msg("attendees limit reached")
		return false
	}
	// Otherwise let request go through
	return true
}
