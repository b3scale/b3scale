package bbb

// An Event is an interface for BBB events.
// All events (that we care about) have a type
// and belong to a meeting.
type Event interface{}

// MeetingCreatedEvent indicates the start of a meeting
type MeetingCreatedEvent struct {
	MeetingID         string
	InternalMeetingID string
}

// MeetingEndedEvent indicates the end of a meeting
type MeetingEndedEvent struct {
	InternalMeetingID string
}

// MeetingDestroyedEvent indicates a meeting destroyed
type MeetingDestroyedEvent struct {
	InternalMeetingID string
}

// UserJoinedMeetingEvent indicates that a user joined the meeting
type UserJoinedMeetingEvent struct {
	InternalMeetingID string
	Attendee          *Attendee
}

// UserLeftMeetingEvent indicates that a user has left the meeting
type UserLeftMeetingEvent struct {
	InternalMeetingID string
	InternalUserID    string
	InternalID        string
}

// RecordingStatusEvent updates the recording status of a meeting
type RecordingStatusEvent struct {
	InternalMeetingID string
	InternalUserID    string
	Recording         bool
}

// BreakoutRoomStartedEvent indicates the start of a breakout room
type BreakoutRoomStartedEvent struct {
	ParentInternalMeetingID string
	Breakout                *BreakoutInfo
}

// BreakoutInfo contains breakout room information
type BreakoutInfo struct {
	Name       string
	ExternalID string
	BreakoutID string
	Sequence   int
	FreeJoin   bool
}
