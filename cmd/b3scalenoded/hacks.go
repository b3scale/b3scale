package main

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"

	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// Errors
var (
	ErrDeadlineReached = errors.New("meeting was not found within the deadline")
)

// Sometime we have to be aware, that the meeting
// might not yet be in the scaler state - however,
// bbb already fired an event for this.
//
// As all events are processed in their own goroutine,
// we sit back, wait, and poll the database
// for the internal meeting to come up. We should give
// up after 2-5 seconds.
func awaitInternalMeeting(
	ctx context.Context,
	internalID string,
	deadlineAfter time.Duration,
) (*store.MeetingState, error) {
	t0 := time.Now()
	for {
		tx, err := store.Begin(ctx)
		if err != nil {
			time.Sleep(150 * time.Millisecond)
			continue
		}
		defer tx.Rollback(ctx)

		mstate, err := store.GetMeetingState(ctx, tx, store.Q().
			Where("meetings.internal_id = ?", internalID))
		if err != nil {
			return nil, err
		}
		if mstate != nil {
			return mstate, nil
		}

		tx.Rollback(ctx) // Close transaction

		dt := time.Now().Sub(t0)
		if dt > deadlineAfter {
			return nil, ErrDeadlineReached
		}

		// Okay we should wait and try again..
		time.Sleep(150 * time.Millisecond)
	}
}
