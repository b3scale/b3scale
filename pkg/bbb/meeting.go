package bbb

import (
	"encoding/xml"
	"fmt"
)

// Breakout info
type Breakout struct {
	XMLName         xml.Name `xml:"breakout" json:"-"`
	ParentMeetingID string   `xml:"parentMeetingID"`
	Sequence        int      `xml:"sequence"`
	FreeJoin        bool     `xml:"freeJoin"`
}

// Attendee of a meeting
type Attendee struct {
	XMLName         xml.Name `xml:"attendee" json:"-"`
	UserID          string   `xml:"userID"`
	InternalUserID  string   `xml:"internalUserID,omit"`
	FullName        string   `xml:"fullName"`
	Role            string   `xml:"role"`
	IsPresenter     bool     `xml:"isPresenter"`
	IsListeningOnly bool     `xml:"isListeningOnly"`
	HasJoinedVoice  bool     `xml:"hasJoinedVoice"`
	HasVideo        bool     `xml:"hasVideo"`
	ClientType      string   `xml:"clientType"`
}

// Meeting information
type Meeting struct {
	XMLName               xml.Name  `xml:"meeting" json:"-"`
	MeetingName           string    `xml:"meetingName"`
	MeetingID             string    `xml:"meetingID"`
	InternalMeetingID     string    `xml:"internalMeetingID"`
	CreateTime            Timestamp `xml:"createTime"`
	CreateDate            string    `xml:"createDate"`
	VoiceBridge           string    `xml:"voiceBridge"`
	DialNumber            string    `xml:"dialNumber"`
	AttendeePW            string    `xml:"attendeePW"`
	ModeratorPW           string    `xml:"moderatorPW"`
	Running               bool      `xml:"running"`
	Duration              int       `xml:"duration"`
	Recording             bool      `xml:"recording"`
	HasBeenForciblyEnded  bool      `xml:"hasBeenForciblyEnded"`
	StartTime             Timestamp `xml:"startTime"`
	EndTime               Timestamp `xml:"endTime"`
	ParticipantCount      int       `xml:"participantCount"`
	ListenerCount         int       `xml:"listenerCount"`
	VoiceParticipantCount int       `xml:"voiceParticipantCount"`
	VideoCount            int       `xml:"videoCount"`
	MaxUsers              int       `xml:"maxUsers"`
	ModeratorCount        int       `xml:"moderatorCount"`
	IsBreakout            bool      `xml:"isBreakout"`

	Metadata Metadata `xml:"metadata"`

	Attendees     []*Attendee `xml:"attendees>attendee"`
	BreakoutRooms []string    `xml:"breakoutRooms>breakout"`
	Breakout      *Breakout   `xml:"breakout"`
}

func (m *Meeting) String() string {
	return fmt.Sprintf(
		"[Meeting id: %v, pc: %v, mc: %v, running: %v]",
		m.MeetingID, m.ParticipantCount, m.ModeratorCount, m.Running,
	)
}

// Update the meeting info with new data
func (m *Meeting) Update(update *Meeting) error {
	if m.MeetingID != update.MeetingID {
		return fmt.Errorf("meeting ids do not match for update")
	}
	if m.InternalMeetingID != update.InternalMeetingID {
		return fmt.Errorf("internal ids do not match for update")
	}
	/*

		if len(update.MeetingName) > 0 {
			m.MeetingName = update.MeetingName
		}
		if len(update.CreateDate) > 0 {
			m.CreateDate = update.CreateDate
		}
		if len(update.VoiceBridge) > 0 {
			m.VoiceBridge = update.VoiceBridge
		}
		if len(update.DialNumber) > 0 {
			m.DialNumber = update.DialNumber
		}
		if len(update.AttendeePW) > 0 {
			m.AttendeePW = update.AttendeePW
		}
		if len(update.ModeratorPW) > 0 {
			m.ModeratorPW = update.ModeratorPW
		}
		m.Running = update.Running
		m.Duration = update.Duration
		m.Recording = update.Recording
		m.HasBeenForciblyEnded = update.HasBeenForciblyEnded
		m.StartTime = update.StartTime
		m.EndTime = update.EndTime
		m.ParticipantCount = update.ParticipantCount
		m.ListenerCount = update.ListenerCount
		m.VoiceParticipantCount = update.VoiceParticipantCount
		m.VideoCount = update.VideoCount
		m.MaxUsers = update.MaxUsers
		m.ModeratorCount = update.ModeratorCount
		m.IsBreakout = update.IsBreakout
		m.Attendees = update.Attendees
		m.BreakoutRooms = update.BreakoutRooms
	*/

	*m = *update

	return nil
}
