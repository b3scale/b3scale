package main

import (
	"log"

	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
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
	return nil
}

// handle event: MeetingEnded
func (h *EventHandler) onMeetingEnded(
	e *bbb.MeetingEndedEvent,
) error {
	log.Println("meeting ended:", e.InternalMeetingID)
	return nil
}

// handle event: MeetingDestroyed
func (h *EventHandler) onMeetingDestroyed(
	e *bbb.MeetingDestroyedEvent,
) error {
	log.Println("meeting destroyed:", e.InternalMeetingID)
	return nil
}

// handle event: UserJoinedMeeting
func (h *EventHandler) onUserJoinedMeeting(
	e *bbb.UserJoinedMeetingEvent,
) error {
	log.Println(
		"user:", e.Attendee.InternalUserID,
		e.Attendee.FullName,
		"joined meeting:", e.InternalMeetingID)
	return nil
}

// handle event: UserLeftMeeting
func (h *EventHandler) onUserLeftMeeting(
	e *bbb.UserLeftMeetingEvent,
) error {
	log.Println(
		"user:", e.InternalUserID,
		"left meeting:", e.InternalMeetingID)
	return nil
}
