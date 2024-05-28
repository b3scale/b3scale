package bbb

import (
	"testing"
)

// Meetings

func TestMeetingUpdate(t *testing.T) {
	data := readTestResponse("getMeetingsSuccess.xml")
	response, _ := UnmarshalGetMeetingsResponse(data)
	m1 := response.Meetings[0]

	response, _ = UnmarshalGetMeetingsResponse(data)
	m2 := response.Meetings[0]

	m2.Running = false
	m2.ParticipantCount = 5000

	m1.Update(m2)

	if m1.ParticipantCount != 5000 {
		t.Error("unexpected participant count")
	}

	m2.MeetingID = "wrong_meeting_id"
	if err := m1.Update(m2); err == nil {
		t.Error("expected error")
	}
}
