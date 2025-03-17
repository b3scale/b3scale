package requests

import (
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
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
		t.Fatal("decode should have been successful")
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

func TestRewriteUniqueMeetingIDRequest(t *testing.T) {
	req := &bbb.Request{
		Params: bbb.Params{
			bbb.ParamMeetingID: "meetingID23",
		},
		Frontend: &bbb.Frontend{
			Key: "frontend42",
		},
	}

	req1 := rewriteUniqueMeetingIDRequest(req)

	mid, _ := req1.Params.MeetingID()
	if mid == "meetingID23" {
		t.Error("expected a changed meeting ID")
	}
}

func TestRewriteUniqueMeetingIDsRequest(t *testing.T) {
	req := &bbb.Request{
		Params: bbb.Params{
			bbb.ParamMeetingID: "meetingID23,meetingID42",
		},
		Frontend: &bbb.Frontend{
			Key: "frontend42",
		},
	}

	req1 := rewriteUniqueMeetingIDRequest(req)

	mids, _ := req1.Params.MeetingIDs()
	if mids[0] == "meetingID23" {
		t.Error("expected a changed meeting ID")
	}
	t.Log(mids)
}

func TestMaybeRewriteMeeting(t *testing.T) {
	m := &bbb.Meeting{
		MeetingID: "WyJma2V5MSIsIm1pZDEiLCJiM3NjbCJd",
		Breakout: &bbb.Breakout{
			ParentMeetingID: "WyJma2V5MSIsIm1pZDEiLCJiM3NjbCJd",
		},
	}

	maybeRewriteMeeting(m)

	if m.MeetingID != "mid1" {
		t.Error("unexpected meeting ID:", m.MeetingID)
	}

	if m.Breakout.ParentMeetingID != "mid1" {
		t.Error("unexpected parent meeting ID:", m.Breakout.ParentMeetingID)
	}
}

func TestMaybeRewriteMeetingsCollection(t *testing.T) {
	c := []*bbb.Meeting{
		&bbb.Meeting{
			MeetingID: "WyJma2V5MSIsIm1pZDEiLCJiM3NjbCJd",
			Breakout: &bbb.Breakout{
				ParentMeetingID: "WyJma2V5MSIsIm1pZDEiLCJiM3NjbCJd",
			},
		},
		&bbb.Meeting{
			MeetingID: "WyJma2V5MSIsIm1pZDEiLCJiM3NjbCJd",
			Breakout: &bbb.Breakout{
				ParentMeetingID: "WyJma2V5MSIsIm1pZDEiLCJiM3NjbCJd",
			},
		},
	}

	maybeRewriteMeetingsCollection(c)

	for _, m := range c {
		if m.MeetingID != "mid1" {
			t.Error("unexpected meeting ID:", m.MeetingID)
		}

		if m.Breakout.ParentMeetingID != "mid1" {
			t.Error("unexpected parent meeting ID:", m.Breakout.ParentMeetingID)
		}
	}
}

func TestMaybeRewriteRecording(t *testing.T) {
	r := &bbb.Recording{
		MeetingID: "WyJma2V5MSIsIm1pZDEiLCJiM3NjbCJd",
		Metadata: bbb.Metadata{
			"meetingId": "WyJma2V5MSIsIm1pZDEiLCJiM3NjbCJd",
		},
	}

	maybeRewriteRecording(r)

	if r.MeetingID != "mid1" {
		t.Error("unexpected meeting id:", r.MeetingID)
	}
	if r.Metadata["meetingId"] != "mid1" {
		t.Error("unexpected meeting id:", r.Metadata["meetingId"])
	}
}

func TestMaybeRewriteRecordingCollection(t *testing.T) {
	c := []*bbb.Recording{
		&bbb.Recording{
			MeetingID: "WyJma2V5MSIsIm1pZDEiLCJiM3NjbCJd",
			Metadata: bbb.Metadata{
				"meetingId": "WyJma2V5MSIsIm1pZDEiLCJiM3NjbCJd",
			},
		},
	}

	maybeRewriteRecordingsCollection(c)

	r := c[0]
	if r.MeetingID != "mid1" {
		t.Error("unexpected meeting id:", r.MeetingID)
	}
	if r.Metadata["meetingId"] != "mid1" {
		t.Error("unexpected meeting id:", r.Metadata["meetingId"])
	}
}

func TestMaybeDecodeMeetingID(t *testing.T) {
	id := "WyJma2V5MSIsIm1pZDEiLCJiM3NjbCJd"
	id1 := maybeDecodeMeetingID(id)
	if id1 == id {
		t.Error("id1 should be decoded")
	}
	if id1 != "mid1" {
		t.Error("unexpected id:", id1)
	}

	id2 := "wyjma2v"
	id3 := maybeDecodeMeetingID(id2)
	if id2 != id3 {
		t.Error("id should not have been touched")
	}
}

func TestRewriteUniqueMeetingIDResponse(t *testing.T) {
	id := "WyJma2V5MSIsIm1pZDEiLCJiM3NjbCJd"
	res1 := &bbb.JoinResponse{
		MeetingID: id,
		UserID:    "test",
	}
	resRewrite, err := rewriteUniqueMeetingIDResponse(res1)
	if err != nil {
		t.Error(err)
	}
	if resRewrite.(*bbb.JoinResponse).MeetingID != "mid1" {
		t.Error("unexpected meetingID")
	}

}

func TestRewriteUniqueMeetingIDResponseMeeting(t *testing.T) {
	id := "WyJma2V5MSIsIm1pZDEiLCJiM3NjbCJd"
	// MeetingInfo
	res1 := &bbb.GetMeetingInfoResponse{
		Meeting: &bbb.Meeting{
			MeetingID: id,
		},
	}
	resRewrite, err := rewriteUniqueMeetingIDResponse(res1)
	if err != nil {
		t.Error(err)
	}
	if resRewrite.(*bbb.GetMeetingInfoResponse).Meeting.MeetingID != "mid1" {
		t.Error("unexpected meetingID")
	}
}

func TestRewriteUniqueMeetingIDResponseCollection(t *testing.T) {
	id := "WyJma2V5MSIsIm1pZDEiLCJiM3NjbCJd"

	// Collection
	res1 := &bbb.GetMeetingsResponse{
		Meetings: []*bbb.Meeting{
			&bbb.Meeting{
				MeetingID: id,
			},
			&bbb.Meeting{
				MeetingID: "foo",
			},
		},
	}
	resRewrite, err := rewriteUniqueMeetingIDResponse(res1)
	if err != nil {
		t.Error(err)
	}
	if resRewrite.(*bbb.GetMeetingsResponse).Meetings[0].MeetingID != "mid1" {
		t.Error("unexpected meetingID")
	}
	if resRewrite.(*bbb.GetMeetingsResponse).Meetings[1].MeetingID != "foo" {
		t.Error("unexpected meetingID")
	}
}
