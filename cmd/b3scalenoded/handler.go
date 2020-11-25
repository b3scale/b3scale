package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// The EventHandler processes BBB Events and updates
// the cluster state
type EventHandler struct {
	pool *pgxpool.Pool
}

// NewEventHandler creates a new handler instance
// with a database pool
func NewEventHandler(pool *pgxpool.Pool) *EventHandler {
	return &EventHandler{pool: pool}
}

// Dispatch invokes the handler functions on the BBB event
func (h *EventHandler) Dispatch(e bbb.Event) error {
	switch e.(type) {
	case *bbb.MeetingCreatedEvent:
		return h.onMeetingCreated(e.(*bbb.MeetingCreatedEvent))
	case *bbb.MeetingEndedEvent:
		return h.onMeetingEnded(e.(*bbb.MeetingEndedEvent))
	case *bbb.MeetingDestroyedEvent:
		return h.onMeetingDestroyed(e.(*bbb.MeetingDestroyedEvent))

	case *bbb.UserJoinedMeetingEvent:
		return h.onUserJoinedMeeting(e.(*bbb.UserJoinedMeetingEvent))
	case *bbb.UserLeftMeetingEvent:
		return h.onUserLeftMeeting(e.(*bbb.UserLeftMeetingEvent))

	default:
		log.Printf("unhandled event: %T %v", e, e)
	}
	return nil
}

// handle event: MeetingCreated
func (h *EventHandler) onMeetingCreated(
	e *bbb.MeetingCreatedEvent,
) error {
	log.Println("meeting created:", e.MeetingID, e.InternalMeetingID)
	// The meeting should be already known, becuase it was created
	// through the scaled. So we try to get the meeting and update
	// it.
	mstate, err := store.GetMeetingState(h.pool, store.Q().
		Where("internal_id = ?", e.InternalMeetingID))
	if err != nil {
		return err
	}
	if mstate == nil {
		log.Println(
			"WARNING: meeting id", e.InternalMeetingID,
			"is not unknown to the cluster")
	}
	mstate.Meeting.Running = true
	if err := mstate.Save(); err != nil {
		return err
	}

	// Do a meeting recount
	ctx := context.Background()
	qry := `
		UPDATE backends
		   SET meetings_count = (
		   	     SELECT COUNT(1) FROM meetings
			      WHERE meetings.backend_id = backends.id)
		  JOIN meetings
		    ON meetings.backend_id = backends.id
		 WHERE meetings.internal_id = $1
	`
	if _, err := h.pool.Exec(ctx, qry, e.InternalMeetingID); err != nil {
		return err
	}

	return nil
}

// handle event: MeetingEnded
func (h *EventHandler) onMeetingEnded(
	e *bbb.MeetingEndedEvent,
) error {
	log.Println("meeting ended:", e.InternalMeetingID)
	mstate, err := store.GetMeetingState(h.pool, store.Q().
		Where("internal_id = ?", e.InternalMeetingID))
	if err != nil {
		return err
	}
	if mstate == nil {
		log.Println(
			"WARNING: meeting id", e.InternalMeetingID,
			"is not unknown to the cluster")
	}
	// Reset meeting state
	mstate.Meeting.Running = false
	mstate.Meeting.Attendees = []*bbb.Attendee{}
	if err := mstate.Save(); err != nil {
		return err
	}
	return mstate.Save()
}

// handle event: MeetingDestroyed
func (h *EventHandler) onMeetingDestroyed(
	e *bbb.MeetingDestroyedEvent,
) error {
	log.Println("meeting destroyed:", e.InternalMeetingID)
	// Delete meeting state
	return store.DeleteMeetingStateByInternalID(h.pool, e.InternalMeetingID)
}

// handle event: UserJoinedMeeting
func (h *EventHandler) onUserJoinedMeeting(
	e *bbb.UserJoinedMeetingEvent,
) error {
	log.Println(
		"user:", e.Attendee.InternalUserID,
		e.Attendee.FullName,
		"joined meeting:", e.InternalMeetingID)

	// Increment attendees
	ctx := context.Background()
	qry := `
		UPDATE backends
		   SET attendees_count = attendees_count + 1
		  JOIN meetings
		    ON meetings.backend_id = backends.id
		 WHERE meetings.internal_id = $1
	`
	if _, err := h.pool.Exec(ctx, qry, e.InternalMeetingID); err != nil {
		return err
	}

	return nil
}

// handle event: UserLeftMeeting
func (h *EventHandler) onUserLeftMeeting(
	e *bbb.UserLeftMeetingEvent,
) error {
	log.Println(
		"user:", e.InternalUserID,
		"left meeting:", e.InternalMeetingID)

	// Decrement attendees
	ctx := context.Background()
	qry := `
		UPDATE backends
		   SET attendees_count = attendees_count - 1
		  JOIN meetings
		    ON meetings.backend_id = backends.id
		 WHERE meetings.internal_id = $1
		   AND attendees_count >= 1
	`
	if _, err := h.pool.Exec(ctx, qry, e.InternalMeetingID); err != nil {
		return err
	}
	return nil
}
