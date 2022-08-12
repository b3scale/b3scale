package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/events"
	"github.com/b3scale/b3scale/pkg/http/api"
	"github.com/b3scale/b3scale/pkg/store"
)

// StartEventMonitor starts listening to events
func StartEventMonitor(
	ctx context.Context,
	cli api.Client,
	rdb *redis.Client,
	backend *store.BackendState,
) {
	monitor := events.NewMonitor(rdb)
	channel := monitor.Subscribe()
	for ev := range channel {
		// We are handling an event in it's own goroutine
		go func(ev bbb.Event) {
			handler := NewEventHandler(backend)
			eventCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			if err := handler.Dispatch(eventCtx, ev); err != nil {
				log.Error().Err(err).Msg("event handler")
			}
		}(ev)
	}
}

// The EventHandler processes BBB Events and updates
// the cluster state
type EventHandler struct {
	api     api.Client
	backend *store.BackendState
}

// NewEventHandler creates a new handler instance
// with a database pool
func NewEventHandler(
	cli api.Client,
	backend *store.BackendState,
) *EventHandler {
	return &EventHandler{
		api:     cli,
		backend: backend,
	}
}

// Dispatch invokes the handler functions on the BBB event
func (h *EventHandler) Dispatch(ctx context.Context, e bbb.Event) error {
	switch e.(type) {
	case *bbb.MeetingCreatedEvent:
		return h.onMeetingCreated(ctx, e.(*bbb.MeetingCreatedEvent))
	case *bbb.MeetingEndedEvent:
		return h.onMeetingEnded(ctx, e.(*bbb.MeetingEndedEvent))
	case *bbb.MeetingDestroyedEvent:
		return h.onMeetingDestroyed(ctx, e.(*bbb.MeetingDestroyedEvent))

	case *bbb.UserJoinedMeetingEvent:
		return h.onUserJoinedMeeting(ctx, e.(*bbb.UserJoinedMeetingEvent))
	case *bbb.UserLeftMeetingEvent:
		return h.onUserLeftMeeting(ctx, e.(*bbb.UserLeftMeetingEvent))

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
	ctx context.Context,
	e *bbb.MeetingCreatedEvent,
) error {
	log.Info().
		Str("internalMeetingID", e.InternalMeetingID).
		Str("meetingID", e.MeetingID).
		Msg("meeting created")

	meeting, err := h.api.MeetingRetrieve(
		ctx, api.InternalMeetingID(e.InternalMeetingID), url.Values{
			"await": "true",
		})
	if err != nil {
		return err
	}

	// Patch meeting
	update, err := json.Marshal(map[string]interface{}{
		"meeting": map[string]interface{}{
			"Running": true,
		},
	})
	if err != nil {
		return err
	}
	_, err = h.api.MeetingUpdateRaw(
		ctx, api.InternalMeetingID(e.InternalMeetingID), update,
	)
	if err != nil {
		return err
	}

	return nil
}

// handle event: MeetingEnded
func (h *EventHandler) onMeetingEnded(
	ctx context.Context,
	e *bbb.MeetingEndedEvent,
) error {
	log.Info().
		Str("internalMeetingID", e.InternalMeetingID).
		Msg("meeting ended")

	meeting, err := h.api.MeetingRetrieve(
		ctx, api.InternalMeetingID(e.InternalMeetingID), url.Values{
			"await": "true",
		})
	if err != nil {
		return err
	}

	// Reset meeting state
	update, err := json.Marshal(map[string]interface{}{
		"meeting": map[string]interface{}{
			"Running":   false,
			"Attendees": []map[string]interface{}{},
		},
	})
	if err != nil {
		return err
	}
	_, err = h.api.MeetingUpdateRaw(
		ctx, api.InternalMeetingID(e.InternalMeetingID), update,
	)
	if err != nil {
		return err
	}

	return nil
}

// handle event: MeetingDestroyed
func (h *EventHandler) onMeetingDestroyed(
	ctx context.Context,
	e *bbb.MeetingDestroyedEvent,
) error {
	log.Info().
		Str("internalMeetingID", e.InternalMeetingID).
		Msg("meeting destroyed")
	// Delete meeting state
	_, err := h.api.MeetingDestroy(
		ctx, api.InternalMeetingID(e.InternalMeetingID))
	if err != nil {
		return err
	}
	return nil
}

// handle event: UserJoinedMeeting
func (h *EventHandler) onUserJoinedMeeting(
	ctx context.Context,
	e *bbb.UserJoinedMeetingEvent,
) error {
	log.Info().
		Str("userID", e.Attendee.InternalUserID).
		Str("userFullName", e.Attendee.FullName).
		Str("internalMeetingID", e.InternalMeetingID).
		Msg("user joined meeting")

	meeting, err := h.api.MeetingRetrieve(
		ctx, api.InternalMeetingID(e.InternalMeetingID), url.Values{
			"await": "true",
		})
	if err != nil {
		return err
	}

	// Update state attendees list
	attendees := meeting.Meeting.Attendees
	if attendees == nil {
		attendees = []*bbb.Attendee{}
	}
	attendees = append(attendees, e.Attendee)
	update, err := json.Marshal(map[string]interface{}{
		"meeting": map[string]interface{}{
			"Attendees": attendees,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// handle event: UserLeftMeeting
func (h *EventHandler) onUserLeftMeeting(
	ctx context.Context,
	e *bbb.UserLeftMeetingEvent,
) error {
	log.Info().
		Str("internalUserID", e.InternalUserID).
		Str("internalMeetingID", e.InternalMeetingID).
		Msg("user left meeting")

	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Remove user from attendees list
	mstate, err := store.GetMeetingState(ctx, tx, store.Q().
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
	if err := mstate.Save(ctx, tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
