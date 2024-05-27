package bbb

import "testing"

func TestUnmarshalRecordingMetadata(t *testing.T) {
	data := readTestResponse("../recordings/metadata.xml")
	res, err := UnmarshalRecordingMetadata(data)
	if err != nil {
		t.Fatal(err)
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

func TestRecordingMetadataToRecording(t *testing.T) {
	data := readTestResponse("../recordings/metadata.xml")
	res, err := UnmarshalRecordingMetadata(data)
	if err != nil {
		t.Fatal(err)
	}

	rec := res.ToRecording()
	if rec.RecordID != "record-2342" {
		t.Error("unexpected recordID:", rec.RecordID)
	}
	if rec.MeetingID != "meeting-2342" {
		t.Error("unexpected meetingID:", rec.MeetingID)
	}
	if rec.InternalMeetingID != "DEPRECATED:meeting-2342" {
		t.Error("unexpected internal meetingID:", rec.InternalMeetingID)
	}

}

func TestRecordingMetadataWithoutMeetingToRecording(t *testing.T) {
	data := readTestResponse("../recordings/metadata.m4v.xml")
	res, err := UnmarshalRecordingMetadata(data)
	if err != nil {
		t.Fatal(err)
	}

	rec := res.ToRecording()
	if rec.RecordID != "record-2342" {
		t.Error("unexpected recordID:", rec.RecordID)
	}
	if rec.MeetingID != "meeting-2342" {
		t.Error("unexpected meetingID:", rec.MeetingID)
	}
	if rec.InternalMeetingID != "DEPRECATED:meeting-2342" {
		t.Error("unexpected internal meetingID:", rec.InternalMeetingID)
	}

}
