package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/bbb"
)

// Errors
var (
	ErrChannelClosed = errors.New("subscription channel disconnected")
)

// A Monitor is connected to a redis server and is
// listening for BBB events.
type Monitor struct {
	rdb *redis.Client
}

// NewMonitor creates a new monitor with a redis connection
func NewMonitor(rdb *redis.Client) *Monitor {
	return &Monitor{
		rdb: rdb,
	}
}

// Subscribe subscribes to the redis store and
// retrieves messsges. These are decoded and returned
// through a channel.
func (m *Monitor) Subscribe() chan bbb.Event {
	events := make(chan bbb.Event)
	go func(events chan bbb.Event) {
		ctx := context.Background()
		pubsub := m.rdb.PSubscribe(ctx, "*akka-apps-redis-channel")
		for {
			err := receiveMessages(events, pubsub)
			log.Error().
				Err(err).
				Msg("redis error on receiveMessages")
			time.Sleep(1 * time.Second)
		}
	}(events)
	return events
}

func receiveMessages(events chan bbb.Event, sub *redis.PubSub) error {
	ctx := context.Background()
	if _, err := sub.Receive(ctx); err != nil {
		return err
	}

	for msg := range sub.Channel() {
		// Decode message and push event
		event := decodeEvent(msg)
		if event == nil {
			continue // We do not really care.
		}
		events <- event
	}

	return ErrChannelClosed
}

// Decode incoming message into a BBB event
func decodeEvent(msg *redis.Message) bbb.Event {
	m := &Message{}
	if err := json.Unmarshal([]byte(msg.Payload), m); err != nil {
		log.Error().
			Err(err).
			Str("data", msg.Payload).
			Msg("decoding eveng")
		return nil
	}

	// Decode event body
	switch m.Envelope.Name {
	case "MeetingCreatedEvtMsg":
		return safeDecode(decodeMeetingCreatedEvent, m)
	case "MeetingEndedEvtMsg":
		return safeDecode(decodeMeetingEndedEvent, m)
	case "MeetingDestroyedEvtMsg":
		return safeDecode(decodeMeetingDestroyedEvent, m)
	case "UserJoinedMeetingEvtMsg":
		return safeDecode(decodeUserJoinedMeetingEvent, m)
	case "UserLeftMeetingEvtMsg":
		return safeDecode(decodeUserLeftMeetingEvent, m)
	case "SetRecordingStatusCmdMsg":
		return safeDecode(decodeSetRecordingStatusEvent, m)
	}

	return nil
}

type decoderFunc func(m *Message) bbb.Event

func safeDecode(decoder decoderFunc, m *Message) bbb.Event {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from:", r)
		}
	}()
	return decoder(m)
}

func decodeMeetingCreatedEvent(m *Message) bbb.Event {
	// Decode "props"
	props := m.Core.Body["props"].(map[string]interface{})
	mprops := props["meetingProp"].(map[string]interface{})
	return &bbb.MeetingCreatedEvent{
		MeetingID:         mprops["extId"].(string),
		InternalMeetingID: mprops["intId"].(string),
	}
}

func decodeMeetingEndedEvent(m *Message) bbb.Event {
	return &bbb.MeetingEndedEvent{
		InternalMeetingID: m.Core.Body["meetingId"].(string),
	}
}

func decodeMeetingDestroyedEvent(m *Message) bbb.Event {
	return &bbb.MeetingDestroyedEvent{
		InternalMeetingID: m.Core.Body["meetingId"].(string),
	}
}

func decodeUserJoinedMeetingEvent(m *Message) bbb.Event {
	user := m.Core.Body
	meetingID := m.Core.Header["meetingId"].(string)
	return &bbb.UserJoinedMeetingEvent{
		InternalMeetingID: meetingID,
		Attendee: &bbb.Attendee{
			UserID:         user["extId"].(string),
			InternalUserID: user["intId"].(string),
			FullName:       user["name"].(string),
			Role:           user["role"].(string),
			ClientType:     user["clientType"].(string),
		},
	}
}

func decodeUserLeftMeetingEvent(m *Message) bbb.Event {
	header := m.Core.Header
	return &bbb.UserLeftMeetingEvent{
		InternalMeetingID: header["meetingId"].(string),
		InternalUserID:    header["userId"].(string),
	}
}

func decodeSetRecordingStatusEvent(m *Message) bbb.Event {
	header := m.Core.Header
	body := m.Core.Body
	return &bbb.RecordingStatusEvent{
		InternalMeetingID: header["meetingId"].(string),
		InternalUserID:    header["userId"].(string),
		Recording:         body["recording"].(bool),
	}
}
