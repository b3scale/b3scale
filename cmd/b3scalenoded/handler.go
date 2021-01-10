package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// The EventHandler processes BBB Events and updates
// the cluster state
type EventHandler struct {
	backend *store.BackendState
	pool    *pgxpool.Pool
}

// NewEventHandler creates a new handler instance
// with a database pool
func NewEventHandler(pool *pgxpool.Pool, backend *store.BackendState) *EventHandler {
	return &EventHandler{
		backend: backend,
		pool:    pool,
	}
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
		log.Error().
			Str("type", fmt.Sprintf("%T", e)).
			Str("event", fmt.Sprintf("%v", e)).
			Msg("unhandled unknown event")
	}
	return nil
}

// handle event: MeetingCreated
func (h *EventHandler) onMeetingCreated(
	e *bbb.MeetingCreatedEvent,
) error {
	log.Info().
		Str("internalMeetingID", e.InternalMeetingID).
		Str("meetingID", e.MeetingID).
		Msg("meeting created")

	deadline := 5 * time.Second
	mstate, err := awaitInternalMeeting(h.pool, e.InternalMeetingID, deadline)
	if err != nil {
		return err
	}
	if mstate == nil {
		log.Warn().
			Str("internalMeetingID", e.InternalMeetingID).
			Str("meetingID", e.MeetingID).
			Msg("meeting identified by internalMeetingID " +
				"is unknown  to the cluster")
		return nil
	}
	mstate.Meeting.Running = true
	if err := mstate.Save(); err != nil {
		return err
	}

	return nil
}

// handle event: MeetingEnded
func (h *EventHandler) onMeetingEnded(
	e *bbb.MeetingEndedEvent,
) error {
	log.Info().
		Str("internalMeetingID", e.InternalMeetingID).
		Msg("meeting ended")
	deadline := 5 * time.Second
	mstate, err := awaitInternalMeeting(h.pool, e.InternalMeetingID, deadline)
	if err != nil {
		return err
	}
	if mstate == nil {
		log.Warn().
			Str("internalMeetingID", e.InternalMeetingID).
			Msg("meeting identified by internalMeetingID " +
				"is unknown to the cluster")
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
	log.Info().
		Str("internalMeetingID", e.InternalMeetingID).
		Msg("meeting destroyed")

	// Delete meeting state
	err := store.DeleteMeetingStateByInternalID(h.pool, e.InternalMeetingID)
	if err != nil {
		return nil
	}
	return nil
}

// handle event: UserJoinedMeeting
func (h *EventHandler) onUserJoinedMeeting(
	e *bbb.UserJoinedMeetingEvent,
) error {
	log.Info().
		Str("userID", e.Attendee.InternalUserID).
		Str("userFullName", e.Attendee.FullName).
		Str("internalMeetingID", e.InternalMeetingID).
		Msg("user joined meeting")

	// Increment attendees
	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second)
	defer cancel()
	qry := `
		UPDATE backends
		   SET attendees_count = attendees_count + 1
		  FROM meetings
		 WHERE meetings.internal_id = $1
		   AND meetings.backend_id = backends.id
	`
	if _, err := h.pool.Exec(ctx, qry, e.InternalMeetingID); err != nil {
		return err
	}

	// Insert (preliminar) attendee into the meeting state
	mstate, err := store.GetMeetingState(h.pool, store.Q().
		Where("meetings.internal_id = ?", e.InternalMeetingID))
	if err != nil {
		return err
	}

	if mstate == nil {
		log.Warn().
			Str("internalMeetingID", e.InternalMeetingID).
			Msg("meeting identified by internalMeetingID " +
				"is unknown to the cluster")
		return nil // however we are done here
	}

	// Update state attendees list
	if mstate.Meeting.Attendees == nil {
		mstate.Meeting.Attendees = []*bbb.Attendee{}
	}
	mstate.Meeting.Attendees = append(
		mstate.Meeting.Attendees,
		e.Attendee)

	if err := mstate.Save(); err != nil {
		return err
	}

	return nil
}

// handle event: UserLeftMeeting
func (h *EventHandler) onUserLeftMeeting(
	e *bbb.UserLeftMeetingEvent,
) error {
	log.Info().
		Str("internalUserID", e.InternalUserID).
		Str("internalMeetingID", e.InternalMeetingID).
		Msg("user left meeting")

	// Decrement attendees
	ctx := context.Background()
	qry := `
		UPDATE backends
		   SET attendees_count = attendees_count - 1
		  FROM meetings
		 WHERE meetings.internal_id = $1
		   AND attendees_count >= 1
		   AND meetings.backend_id = backends.id
	`
	if _, err := h.pool.Exec(ctx, qry, e.InternalMeetingID); err != nil {
		return err
	}

	// Remove user from attendees list
	mstate, err := store.GetMeetingState(h.pool, store.Q().
		Where("meetings.internal_id = ?", e.InternalMeetingID))
	if err != nil {
		return err
	}

	if mstate == nil {
		log.Warn().
			Str("internalMeetingID", e.InternalMeetingID).
			Msg("meeting identified by internalMeetingID " +
				"is unknown to the cluster")
		return nil // however we are done here
	}

	// Update state attendees list
	if mstate.Meeting.Attendees == nil {
		return nil // Unlikely but in this case, we are done here
	}

	// Remove user from meeting's attendees
	filtered := make([]*bbb.Attendee, 0, len(mstate.Meeting.Attendees))
	for _, a := range mstate.Meeting.Attendees {
		if a.InternalUserID == e.InternalUserID {
			continue // The user just left
		}
		filtered = append(filtered, a)
	}
	mstate.Meeting.Attendees = filtered

	if err := mstate.Save(); err != nil {
		return err
	}
	return nil
}
