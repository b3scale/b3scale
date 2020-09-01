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

// CreateResponse is the resonse for the `create` API resource.
type CreateResponse struct {
	*XMLResponse
	*Meeting
}

// UnmarshalCreateResponse decodes the resonse XML data.
func UnmarshalCreateResponse(data []byte) (*CreateResponse, error) {
	res := &CreateResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal a CreateResponse to XML
func (res *CreateResponse) Marshal() ([]byte, error) {
	data, err := xml.Marshal(res)
	return data, err
}

// JoinResponse of the join resource
type JoinResponse struct {
	*XMLResponse
	MeetingID    string `xml:"meeting_id"`
	UserID       string `xml:"user_id"`
	AuthToken    string `xml:"auth_token"`
	SessionToken string `xml:"session_token"`
	URL          string `xml:"url"`
}

// UnmarshalJoinResponse decodes the serialized XML data
func UnmarshalJoinResponse(data []byte) (*JoinResponse, error) {
	res := &JoinResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal encodes a JoinResponse as XML
func (res *JoinResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// IsMeetingRunningResponse is a meeting status resonse
type IsMeetingRunningResponse struct {
	*XMLResponse
	Running bool `xml:"running"`
}

// UnmarshalIsMeetingRunningResponse decodes the XML data.
func UnmarshalIsMeetingRunningResponse(
	data []byte,
) (*IsMeetingRunningResponse, error) {
	res := &IsMeetingRunningResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal a IsMeetingRunningResponse to XML
func (res *IsMeetingRunningResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// EndResponse is the resonse of the end resource
type EndResponse struct {
	*XMLResponse
}

// UnmarshalEndResponse decodes the xml resonse
func UnmarshalEndResponse(data []byte) (*EndResponse, error) {
	res := &EndResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal EndResponse to XML
func (res *EndResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// GetMeetingInfoResponse contains detailed meeting information
type GetMeetingInfoResponse struct {
	*XMLResponse
	*Meeting
}

// UnmarshalGetMeetingInfoResponse decodes the xml response
func UnmarshalGetMeetingInfoResponse(
	data []byte,
) (*GetMeetingInfoResponse, error) {
	res := &GetMeetingInfoResponse{} // Meeting: &Meeting{AttendeesCollection: &AttendeesCollection{}}}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal GetMeetingInfoResponse to XML
func (res *GetMeetingInfoResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// GetMeetingsResponse contains a list of meetings.
type GetMeetingsResponse struct {
	*XMLResponse
	Meetings []*Meeting `xml:"data"`
}

// BreakoutRooms is a collection of breakout room ids
type BreakoutRooms struct {
	XMLName     xml.Name `xml:"breakoutRooms"`
	BreakoutIDs []string `xml:"breakout"`
}

// Breakout info
type Breakout struct {
	XMLName         xml.Name `xml:"breakout"`
	ParentMeetingID string   `xml:"parentMeetingID"`
	Sequence        int      `xml:"sequence"`
	FreeJoin        bool     `xml:"freeJoin"`
}

// AttendeesCollection contains a list of attendees
type AttendeesCollection struct {
	XMLName   xml.Name    `xml:"attendees"`
	Attendees []*Attendee `xml:"attendee"`
}

// Attendee of a meeting
type Attendee struct {
	XMLName         xml.Name `xml:"attendee"`
	UserID          string   `xml:"userID"`
	FullName        string   `xml:"fullName"`
	Role            string   `xml:"role"`
	IsPresenter     bool     `xml:"isPresenter"`
	IsListeningOnly bool     `xml:"isListeningOnly"`
	HasJoinedVoice  bool     `xml:"hasJoinedVoice"`
	HasVideo        bool     `xml:"hasVideo"`
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
	MeetingName           string    `xml:"meetingName"`
	MeetingID             string    `xml:"meetingID"`
	InternalMeetingID     string    `xml:"internalMeetingID"`
	CreateTime            Timestamp `xml:"createTime"`
	CreateDate            string    `xml:"createDate"`
	VoiceBridge           string    `xml:"voiceBridge"`
	DialNumber            string    `xml:"dialNumber"`
	AttendeePW            string    `xml:"attendeePW"`
	ModeratorPW           string    `xml:"moderatorPW"`
	Running               string    `xml:"running"`
	Duration              int       `xml:"duration"`
	Recording             string    `xml:"recording"`
	HasBeenForciblyEnded  string    `xml:"hasBeenForciblyEnded"`
	StartTime             Timestamp `xml:"startTime"`
	EndTime               Timestamp `xml:"endTime"`
	ParticipantCount      int       `xml:"participantCount"`
	ListenerCount         int       `xml:"listenerCount"`
	VoiceParticipantCount int       `xml:"voiceParticipantCount"`
	VideoCount            int       `xml:"videoCount"`
	MaxUsers              int       `xml:"maxUsers"`
	ModeratorCount        int       `xml:"moderatorCount"`
	Metadata              Metadata  `xml:"metadata"`
	IsBreakout            bool      `xml:"isBreakout"`

	AttendeesCollection *AttendeesCollection `xml:"attendees"`
	*BreakoutRooms
	*Breakout

	SyncedAt time.Time
	Mux      sync.Mutex
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
	m.AttendeesCollection = meeting.AttendeesCollection
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
