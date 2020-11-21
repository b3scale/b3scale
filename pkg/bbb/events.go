package bbb

// An Event is an interface for BBB events.
// All events (that we care about) have a type
// and belong to a meeting.
type Event interface{}

// MeetingStartedEvent indicates a meeting start
type MeetingStartedEvent struct {
	MeetingID string
}

// MeetingEndedEvent indicates the end of a meeting
type MeetingEndedEvent struct {
	MeetingID string
}

// MeetingDestroyedEvent indicates a meeting destroyed
type MeetingDestroyedEvent struct {
	MeetingID string
}

// UserJoinedMeetingEvent indicates that a user joined the meeting
type UserJoinedMeetingEvent struct {
	MeetingID   string
	UserID      string
	InternalID  string
	ExternalID  string
	Name        string
	Role        string
	Guest       bool
	Authed      bool
	GuestStatus string
	Emoji       string
	Presenter   bool
	Locked      bool
	Avatar      string
	ClientType  string
}

// UserLeftMeetingEvent indicates that a user has left the meeting
type UserLeftMeetingEvent struct {
	MeetingID  string
	UserID     string
	InternalID string
}

// BreakoutRoomStartedEvent indicates the start of a breakout room
type BreakoutRoomStartedEvent struct {
	ParentMeetingID string
	Breakout        *BreakoutInfo
}

// BreakoutInfo contains breakout room information
type BreakoutInfo struct {
	Name       string
	ExternalID string
	BreakoutID string
	Sequence   int
	FreeJoin   bool
}
