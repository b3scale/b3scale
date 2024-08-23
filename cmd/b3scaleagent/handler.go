package main

import (
	"context"
	"fmt"
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
			handler := NewEventHandler(cli, backend)
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
	case *bbb.RecordingStatusEvent:
		return h.onRecordingStatus(ctx, e.(*bbb.RecordingStatusEvent))

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
	_, err := h.api.AgentRPC(
		ctx, api.RPCMeetingSetRunning(&api.MeetingSetRunningRequest{
			InternalMeetingID: e.InternalMeetingID,
			Running:           true,
		}))
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

	_, err := h.api.AgentRPC(
		ctx, api.RPCMeetingStateReset(&api.MeetingStateResetRequest{
			InternalMeetingID: e.InternalMeetingID,
		}))
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
	_, err := h.api.MeetingDelete(
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

	_, err := h.api.AgentRPC(
		ctx, api.RPCMeetingAddAttendee(&api.MeetingAddAttendeeRequest{
			InternalMeetingID: e.InternalMeetingID,
			Attendee:          e.Attendee,
		}))
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

	_, err := h.api.AgentRPC(
		ctx, api.RPCMeetingRemoveAttendee(&api.MeetingRemoveAttendeeRequest{
			InternalMeetingID: e.InternalMeetingID,
			InternalUserID:    e.InternalUserID,
		}))
	if err != nil {
		return err
	}

	return nil
}

// handle event: RecordingStatusEvent
func (h *EventHandler) onRecordingStatus(
	ctx context.Context,
	e *bbb.RecordingStatusEvent,
) error {
	log.Info().
		Str("meetingId", e.InternalMeetingID).
		Str("userId", e.InternalUserID).
		Bool("recording", e.Recording).
		Msg("recording status")

	_, err := h.api.AgentRPC(
		ctx, api.RPCMeetingUpdateRecordingStatus(&api.MeetingUpdateRecordingStatusRequest{
			InternalMeetingID: e.InternalMeetingID,
			// TODO: Remove?
			InternalUserID: e.InternalUserID,
			Recording:      e.Recording,
		}))
	if err != nil {
		return err
	}

	return nil
}
