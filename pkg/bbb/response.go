package bbb

import (
	"encoding/xml"
	"fmt"
	"sync"
	"time"
)

// A XMLResponse from the server
type XMLResponse struct {
	XMLName    xml.Name `xml:"response"`
	Returncode string   `xml:"returncode"`
	Message    string   `xml:"message"`
	MessageKey string   `xml:"messageKey"`
}

// IsMeetingRunningResponse is a meeting status response
type IsMeetingRunningResponse struct {
	*XMLResponse
	Running bool `xml:"running"`
}

// UnmarshalIsMeetingRunningResponse decodes the XML data.
func UnmarshalIsMeetingRunningResponse(
	data []byte,
) (*IsMeetingRunningResponse, error) {
	resp := &IsMeetingRunningResponse{}
	err := xml.Unmarshal(data, resp)
	return resp, err
}

// Marshal a IsMeetingRunningResponse to XML
func (res *IsMeetingRunningResponse) Marshal() ([]byte, error) {
	data, err := xml.Marshal(res)
	return data, err
}

// GetMeetingsResponse contains a list of meetings.
type GetMeetingsResponse struct {
	*XMLResponse
	Meetings []*Meeting `xml:"data"`
}

// Attendees collection
type Attendees struct {
	XMLName   xml.Name   `xml:"attendees"`
	Attendees []Attendee `xml:"attendee"`
}

// Attendee of a meeting
type Attendee struct {
	XMLName         xml.Name `xml:"attendee"`
	UserID          string   `xml:"userID"`
	FullName        string   `xml:"fullName"`
	IsPresenter     string   `xml:"isPresenter"`
	IsListeningOnly string   `xml:"isListeningOnly"`
	HasJoinedVoice  string   `xml:"hasJoinedVoice"`
	HasVideo        string   `xml:"hasVideo"`
	ClientType      string   `xml:"clientType"`
}

// Metadata about the BBB instance
type Metadata struct {
	XMLName             xml.Name `xml:"metadata"`
	BBBOriginVersion    string   `xml:"bbb-origin-version"`
	BBBOriginServerName string   `xml:"bbb-origin-server-name"`
	BBBOrigin           string   `xml:"bbb-origin"`
	GlListed            string   `xml:"gl-listed"`
}

// Meeting information
type Meeting struct {
	XMLName               xml.Name  `xml:"meeting"`
	MeetingName           string    `xml:"meetingName"`
	MeetingID             string    `xml:"meetingID"`
	InternalMeetingID     string    `xml:"internalMeetingID"`
	CreateTime            uint64    `xml:"createTime"`
	CreateDate            string    `xml:"createDate"`
	VoiceBridge           string    `xml:"voiceBridge"`
	DialNumber            string    `xml:"dialNumber"`
	AttendeePW            string    `xml:"attendeePW"`
	ModeratorPW           string    `xml:"moderatorPW"`
	Running               string    `xml:"running"`
	Duration              int16     `xml:"duration"`
	Recording             string    `xml:"recording"`
	HasBeenForciblyEnded  string    `xml:"hasBeenForciblyEnded"`
	StartTime             uint64    `xml:"startTime"`
	EndTime               uint64    `xml:"endTime"`
	ParticipantCount      uint32    `xml:"participantCount"`
	ListenerCount         uint32    `xml:"listenerCount"`
	VoiceParticipantCount uint32    `xml:"voiceParticipantCount"`
	VideoCount            uint32    `xml:"videoCount"`
	MaxUsers              uint32    `xml:"maxUsers"`
	ModeratorCount        uint32    `xml:"moderatorCount"`
	Attendees             Attendees `xml:"attendees"`
	Metadata              Metadata  `xml:"metadata"`
	IsBreakout            string    `xml:"isBreakout"`

	SyncedAt time.Time
	Mux      sync.Mutex
}

// MeetingInfo contains getMeetingInfo details
type MeetingInfo struct {
	Meeting
	XMLName xml.Name `xml:"response"`
}

// Update meeting fields
func (m *Meeting) Update(meeting *Meeting) {
	m.Mux.Lock()
	defer m.Mux.Unlock()

	// This is kind of ugly but does not warrent including
	// a new dependecy like `mergo`
	m.MeetingName = meeting.MeetingName
	m.InternalMeetingID = meeting.InternalMeetingID
	m.CreateTime = meeting.CreateTime
	m.CreateDate = meeting.CreateDate
	m.VoiceBridge = meeting.VoiceBridge
	m.DialNumber = meeting.DialNumber
	m.AttendeePW = meeting.AttendeePW
	m.ModeratorPW = meeting.ModeratorPW
	m.Running = meeting.Running
	m.Recording = meeting.Recording
	m.Duration = meeting.Duration
	m.HasBeenForciblyEnded = meeting.HasBeenForciblyEnded
	m.StartTime = meeting.StartTime
	m.EndTime = meeting.EndTime
	m.ParticipantCount = meeting.ParticipantCount // This is why we are here
	m.ListenerCount = meeting.ListenerCount
	m.VoiceParticipantCount = meeting.VoiceParticipantCount
	m.VideoCount = meeting.VideoCount
	m.MaxUsers = meeting.MaxUsers
	m.ModeratorCount = meeting.ModeratorCount
	m.Attendees = meeting.Attendees
	m.Metadata = meeting.Metadata
	m.IsBreakout = meeting.IsBreakout

	// Update sync timestamp
	m.SyncedAt = time.Now()
}

func (m *Meeting) String() string {
	return fmt.Sprintf(
		"[Meeting id: %v, pc: %v, mc: %v, running: %v]",
		m.MeetingID, m.ParticipantCount, m.ModeratorCount, m.Running,
	)
}
