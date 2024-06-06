package bbb

import "testing"

func TestUnmarshalGetRecordingTextTracksResponse(t *testing.T) {
	data := readTestResponse("getRecordingTextTracksSuccess.json")
	response, err := UnmarshalGetRecordingTextTracksResponse(data)
	if err != nil {
		t.Error(err)
	}

	tracks := response.Tracks
	if len(tracks) != 2 {
		t.Error("Unexpected Tracks:", tracks)
	}
	if tracks[0].Label != "English" {
		t.Error("Exptected English:", tracks[0].Label)
	}
}

func TestMarshalGetRecordingTextTracksResponse(t *testing.T) {
	data := readTestResponse("getRecordingTextTracksSuccess.json")
	response, err := UnmarshalGetRecordingTextTracksResponse(data)
	if err != nil {
		t.Error(err)
	}
	data1, err := response.Marshal()
	if err != nil {
		t.Error(err)
	}

	if len(data1) != 489 {
		t.Error("Unexpected data:", string(data1), len(data1))
	}
}

func TestUnmarshalPutRecordingTextTrackResponse(t *testing.T) {
	data := readTestResponse("putRecordingTextTrackSuccess.json")
	response, err := UnmarshalPutRecordingTextTrackResponse(data)
	if err != nil {
		t.Error(err)
	}
	if response.RecordID != "baz" {
		t.Error("Unexpected:", response.RecordID)
	}
	if response.Returncode != "SUCCESS" {
		t.Error("Unexpected:", response.Returncode)
	}
}

func TestMarshalPutRecordingTextTrackResponse(t *testing.T) {
	res := PutRecordingTextTrackResponse{
		Returncode: "SUCCESS",
		RecordID:   "y4aaY",
	}
	data, err := res.Marshal()
	if err != nil {
		t.Error(err)
	}
	if len(data) != 56 {
		t.Error("Unexpected:", string(data), len(data))
	}
}

func TestMergeRecordings(t *testing.T) {
	data := readTestResponse("../recordings/metadata.xml")
	meta, err := UnmarshalRecordingMetadata(data)
	if err != nil {
		t.Fatal(err)
	}
	rec1 := meta.ToRecording()

	data = readTestResponse("../recordings/metadata.m4v.xml")
	meta, err = UnmarshalRecordingMetadata(data)
	if err != nil {
		t.Fatal(err)
	}
	rec2 := meta.ToRecording()

	rec2.Merge(rec1)
	if rec2.Name != rec1.Name {
		t.Error("Unexpected Name:", rec2.Name)
	}
	if rec2.RecordID != rec1.RecordID {
		t.Error("Unexpected RecordID:", rec2.RecordID)
	}
	if rec2.MeetingID != rec1.MeetingID {
		t.Error("Unexpected MeetingID:", rec2.MeetingID)
	}

	if len(rec2.Formats) != 2 {
		t.Error("Unexpected formats:", rec2.Formats)
	}

	for _, f := range rec2.Formats {
		t.Log(f)
	}
}

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
