package bbb

import "testing"

func TestUnmarshalRecordingMetadata(t *testing.T) {
	data := readTestResponse("../recordings/metadata.xml")
	res, err := UnmarshalRecordingMetadata(data)
	if err != nil {
		t.Error(err)
	}
	if res.RecordID != "record-2342" {
		t.Error("unexpected recordID:", res.RecordID)
	}
	if res.Meeting.MeetingID != "meeting-2342" {
		t.Error("unexpected meetingID:", res.Meeting.MeetingID)
	}
	if res.Meeting.InternalMeetingID != "internal-2342" {
		t.Error("unexpected internal meetingID:", res.Meeting.InternalMeetingID)
	}
	if res.Playback.Format != "presentation" {
		t.Error("unexpected format:", res.Playback.Format)
	}
}
