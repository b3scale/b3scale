package events

import ()

// An Event is an BBB event of a type and payload
type Event interface{}

// A Message represents an akka message on the channel
type Message struct {
	Envelope *MessageEnvelope `json:"envelope"`
	Core     *MessageCore     `json:"core"`
}

// MessageEnvelope of the akka message
type MessageEnvelope struct {
	Name      string            `json:"name"`
	Routing   map[string]string `json:"routing"`
	Timestamp int               `json:"timestamp"`
}

// The MessageCore is the 'core' of the message
type MessageCore struct {
	Header map[string]string
	Body   map[string]interface{}
}
