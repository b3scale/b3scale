package events

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
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
// retrievs messsges. These are decoded and returned
// through a channel.
func (m *Monitor) Subscribe() chan Event {
	events := make(chan Event)
	go func(events chan Event) {
		ctx := context.Background()
		pubsub := m.rdb.PSubscribe(ctx, "*akka-apps-redis-channel")
		for {
			err := receiveMessages(events, pubsub)
			log.Println("redis error:", err)
			time.Sleep(1 * time.Second)
		}
	}(events)
	return events
}

func receiveMessages(events chan Event, sub *redis.PubSub) error {
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
func decodeEvent(msg *redis.Message) Event {
	m := &Message{}
	if err := json.Unmarshal([]byte(msg.Payload), m); err != nil {
		log.Println(err)
		return nil
	}

	return m
}
