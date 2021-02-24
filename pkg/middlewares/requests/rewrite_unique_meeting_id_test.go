package requests

import (
	"testing"
)

func TestFrontendKeyMeetingIDEncodeDecode(t *testing.T) {
	id := &FrontendKeyMeetingID{
		FrontendKey: "fkey1",
		MeetingID:   "mid1",
	}

	enc := id.EncodeToString()
	t.Log(enc)

	id2 := DecodeFrontendKeyMeetingID(enc)
	if id2 == nil {
		t.Error("decode should have been successful")
	}

	if id2.FrontendKey != "fkey1" {
		t.Error("unexpected frontend key:", id2.FrontendKey)
	}
	if id2.MeetingID != "mid1" {
		t.Error("unexpected meetingID:", id2.MeetingID)
	}
}

func TestFrontendKeyMeetingIDEncodeDecodeError(t *testing.T) {
	id := "WyJma2V5MSIsIm1pZDEiLCJiM3NjcCJd"
	decode := DecodeFrontendKeyMeetingID(id)
	if decode != nil {
		t.Error("decode should have been unsuccessful")
	}
}
