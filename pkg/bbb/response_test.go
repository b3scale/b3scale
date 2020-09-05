package bbb

import (
	"io/ioutil"
	"path"
	"testing"
)

func readTestResponse(name string) []byte {
	filename := path.Join(
		"../../testdata/responses/",
		name)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return data
}

func TestUnmarshalCreateResponse(t *testing.T) {
	data := readTestResponse("createSuccess.xml")
	response, err := UnmarshalCreateResponse(data)
	if err != nil {
		t.Error(err)
	}
	if response.MeetingID != "Test" {
		t.Error("Unexpected meeting id")
	}
}

func TestMarshalCreateResponse(t *testing.T) {
	data := readTestResponse("createSuccess.xml")
	response, err := UnmarshalCreateResponse(data)
	if err != nil {
		t.Error(err)
	}

	data1, err := response.Marshal()
	if err != nil {
		t.Error(err)
	}
	if len(data1) != 1039 {
		t.Error("Unexpected data:", string(data1), len(data1))
	}
}

func TestUnmarshalJoinResponse(t *testing.T) {
	data := readTestResponse("joinSuccess.xml")
	response, err := UnmarshalJoinResponse(data)
	if err != nil {
		t.Error(err)
	}
	if response.UserID != "w_euxnssffnsbs" {
		t.Error("Unexpected UserID:", response.UserID)
	}
}

func TestMarshalJoinResponse(t *testing.T) {
	data := readTestResponse("joinSuccess.xml")
	response, err := UnmarshalJoinResponse(data)
	if err != nil {
		t.Error(err)
	}
	// Marshal and check result
	data1, err := response.Marshal()
	if err != nil {
		t.Error(err)
	}
	if len(data1) < 80 {
		t.Error("Unexpecetd data:", data1)
	}
}

func TestUnmarshalIsMeetingRunningResponse(t *testing.T) {
	data := readTestResponse("isMeetingRunningSuccess.xml")
	response, err := UnmarshalIsMeetingRunningResponse(data)
	if err != nil {
		t.Error(err)
	}

	if response.XMLResponse.Returncode != "SUCCESS" {
		t.Error("Unexpected returncode:",
			response.XMLResponse.Returncode)
	}
	if response.Running != true {
		t.Error("Expected running to be true")
	}
}

func TestMarshalIsMeetingRunningResponse(t *testing.T) {
	data := readTestResponse("isMeetingRunningSuccess.xml")
	response, err := UnmarshalIsMeetingRunningResponse(data)
	if err != nil {
		t.Error(err)
	}
	data1, err := response.Marshal()
	if err != nil {
		t.Error(err)
	}
	if len(data1) != 120 {
		t.Error("Unexpected data:", string(data1), len(data1))
	}
}

func TestUnmarshalEndResponse(t *testing.T) {
	data := readTestResponse("endSuccess.xml")
	response, err := UnmarshalEndResponse(data)
	if err != nil {
		t.Error(err)
	}

	if response.XMLResponse.MessageKey != "sentEndMeetingRequest" {
		t.Error("Unexpected MessageKey:", response.XMLResponse.MessageKey)
	}
}

func TestUnmarshalGetMeetingInfoRespons(t *testing.T) {
	data := readTestResponse("getMeetingInfoSuccess.xml")
	response, err := UnmarshalGetMeetingInfoResponse(data)
	if err != nil {
		t.Error(err)
	}

	if response.Meeting.MeetingID != "Demo Meeting" {
		t.Error("Unexpected MeetingID:", response.Meeting.MeetingID)
	}

	if response.Meeting.Attendees == nil {
		t.Error("Attendees is nil")
	}

	if len(response.Meeting.Attendees) != 2 {
		t.Error("Unexpected attendees length:",
			len(response.Meeting.Attendees))
	}

	if len(response.Meeting.Metadata) != 0 {
		t.Error("Unexpected Meta:", response.Meeting.Metadata)
	}
}

func TestUnmarshalGetMeetingInfoResponsBreakout(t *testing.T) {
	// We have a breakout room response
	data := readTestResponse("getMeetingInfoSuccess-breakout.xml")
	response, err := UnmarshalGetMeetingInfoResponse(data)
	if err != nil {
		t.Error(err)
	}

	if response.Meeting.MeetingID != "Demo Meeting" {
		t.Error("Unexpected MeetingID:", response.Meeting.MeetingID)
	}

	if response.Meeting.IsBreakout == false {
		t.Error("Should be a breakout room")
	}

	if response.Meeting.Breakout.ParentMeetingID != "ParentMeetingId" {
		t.Error("Unexpected parentMeetingId:",
			response.Meeting.Breakout.ParentMeetingID)
	}
}

func TestUnmarshalGetMeetingInfoResponsBreakoutParent(t *testing.T) {
	// We have a breakout parent response
	data := readTestResponse("getMeetingInfoSuccess-breakoutParent.xml")
	response, err := UnmarshalGetMeetingInfoResponse(data)
	if err != nil {
		t.Error(err)
	}
	if response.Meeting.MeetingID != "Demo Meeting" {
		t.Error("Unexpected MeetingID:", response.Meeting.MeetingID)
	}

	if response.Meeting.IsBreakout == true {
		t.Error("Should not be a breakout room")
	}

	// Check breakout info
	rooms := response.Meeting.BreakoutRooms
	if len(rooms) != 3 {
		t.Error("Expected 3 breakout ids. got:", len(rooms))
	}

	if rooms[0] != "breakout-room-id-1" {
		t.Error("Unexpected breakout ID:",
			rooms[0])
	}
}

func TestMarshalGetMeetingInfoResponse(t *testing.T) {
	data := readTestResponse("getMeetingInfoSuccess.xml")
	response, err := UnmarshalGetMeetingInfoResponse(data)
	if err != nil {
		t.Error(err)
	}
	// Serialize
	data1, err := response.Marshal()
	if err != nil {
		t.Error(err)
	}

	if len(data1) != 1486 {
		t.Error("Unexpected data:", string(data1), len(data1))
	}
}

func TestUnmarshalGetMeetingsResponse(t *testing.T) {
	data := readTestResponse("getMeetingsSuccess.xml")
	response, err := UnmarshalGetMeetingsResponse(data)
	if err != nil {
		t.Error(err)
	}

	if len(response.Meetings) != 1 {
		t.Error("Expected 1 meeting, got:",
			len(response.Meetings))
	}
	meeting := response.Meetings[0]
	if meeting.MeetingID != "Demo Meeting" {
		t.Error("Unexpected MeetingID",
			meeting.MeetingID)
	}
}

func TestMarshalGetMeetingsResponse(t *testing.T) {
	data := readTestResponse("getMeetingsSuccess.xml")
	response, err := UnmarshalGetMeetingsResponse(data)
	if err != nil {
		t.Error(err)
	}
	data1, err := response.Marshal()
	if err != nil {
		t.Error(err)
	}
	if len(data1) != 1005 {
		t.Error("Unexpected data:", len(data1))
		t.Log(string(data1))
	}
}

func TestUnmarshalGetRecordingsResponse(t *testing.T) {
	data := readTestResponse("getRecordingsSuccess.xml")
	response, err := UnmarshalGetRecordingsResponse(data)
	if err != nil {
		t.Error(err)
	}

	recordings := response.Recordings
	if len(recordings) != 2 {
		t.Error("Unexpected recordings:", recordings)
	}

	if recordings[0].Metadata["meetingName"] != "Fred's Room" {
		t.Error("Unexpected metadata:", recordings[0].Metadata)
	}

	if len(recordings[0].Formats) != 2 {
		t.Error("Unexpected formats:", recordings[0].Formats)
	}
}

func TestMarshalGetRecordingsResponse(t *testing.T) {
	data := readTestResponse("getRecordingsSuccess.xml")
	response, err := UnmarshalGetRecordingsResponse(data)
	if err != nil {
		t.Error(err)
	}
	// Marshal response
	data1, err := response.Marshal()
	if err != nil {
		t.Error(err)
	}
	if len(data1) != 2911 {
		t.Error("Unexpected data:", string(data1), len(data1))
	}
}

func TestUnmarshalPublishRecordingsResponse(t *testing.T) {
	data := readTestResponse("publishRecordingsSuccess.xml")
	response, err := UnmarshalPublishRecordingsResponse(data)
	if err != nil {
		t.Error(err)
	}
	if response.Published != true {
		t.Error("Expected recordings to be published")
	}
}

func TestMarshalPublishRecordingsResponse(t *testing.T) {
	data := readTestResponse("publishRecordingsSuccess.xml")
	response, err := UnmarshalPublishRecordingsResponse(data)
	if err != nil {
		t.Error(err)
	}
	data1, err := response.Marshal()
	if len(data1) != 124 {
		t.Error("Unexpected data:", string(data1), len(data1))
	}
}
