package events

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

// The MessageCore is a "core-message"
type MessageCore struct {
	Header map[string]string      `json:"header"`
	Body   map[string]interface{} `json:"body"`
}
