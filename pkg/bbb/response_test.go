package bbb

import (
	"io/ioutil"
	"path"
	"testing"
)

func readTestResponse(name string) []byte {
	filename := path.Join(
		"../../test/data/responses/",
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
	t.Log(string(data1))
}

func TestUnmarshalJoinResponse(t *testing.T) {
	data := readTestResponse("joinSuccess.xml")
	response, err := UnmarshalJoinResponse(data)
	if err != nil {
		t.Error(err)
	}
	t.Log(response)
}

func TestUnmarshalIsMeetingRunningResponse(t *testing.T) {
	data := readTestResponse("isMeetingRunningSuccess.xml")
	response, err := UnmarshalIsMeetingRunningResponse(data)
	if err != nil {
		t.Error(err)
	}

	if response.XMLResponse.Returncode != "SUCCESS" {
		t.Error("Unexpected returncode:", response.XMLResponse.Returncode)
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
	t.Log(string(data1))
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
	t.Log(response)

	if response.Meeting.MeetingID != "Demo Meeting" {
		t.Error("Unexpected MeetingID:", response.Meeting.MeetingID)
	}

	if response.Meeting.AttendeesCollection == nil {
		t.Error("AttendeesCollection is nil")
	}

	if len(response.Meeting.AttendeesCollection.Attendees) != 2 {
		t.Error("Unexpected attendees length:",
			len(response.Meeting.AttendeesCollection.Attendees))
	}
}
