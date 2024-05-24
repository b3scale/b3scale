package api

import (
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/store"
)

func testRPCRequest(t *testing.T, req *RPCRequest) interface{} {
	api, res := NewTestRequest().
		KeepState().
		Authorize("test-agent-2000", auth.ScopeNode).
		JSON(req).
		Context()
	defer api.Release()

	// Call API
	if err := api.Handle(ResourceAgentRPC.Create); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}
	body := res.JSON()
	if body["status"].(string) != RPCStatusOK {
		t.Error("unexpected status", body)
	}
	return body
}

func TestMeetingStateReset(t *testing.T) {
	api, _ := NewTestRequest().Context()
	defer api.Release()

	backend := createTestBackend(api)
	meeting := createTestMeeting(api, backend)

	if meeting.Meeting.Running == false {
		t.Fatal("unexpected meeting:", meeting)
	}

	// Make request
	rpc := RPCMeetingStateReset(&MeetingStateResetRequest{
		InternalMeetingID: meeting.InternalID,
	})
	testRPCRequest(t, rpc)

	// Get Meeting and check if running was resetted
	tx, err := api.Conn.Begin(api.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(api.Ctx())
	state, err := store.GetMeetingState(api.Ctx(), tx, store.Q().
		Where("meetings.id = ?", meeting.ID))
	if err != nil {
		t.Error(err)
	}
	if state.Meeting.Running != false {
		t.Error("unexpected meeting state:", state)
	}
}

func TestMeetingSetRunning(t *testing.T) {
	api, _ := NewTestRequest().Context()
	defer api.Release()

	backend := createTestBackend(api)
	meeting := createTestMeeting(api, backend)

	// Make request
	rpc := RPCMeetingSetRunning(&MeetingSetRunningRequest{
		InternalMeetingID: meeting.InternalID,
		Running:           false,
	})
	testRPCRequest(t, rpc)

	// Get Meeting and check if running was resetted
	tx, err := api.Conn.Begin(api.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(api.Ctx())
	state, err := store.GetMeetingState(api.Ctx(), tx, store.Q().
		Where("meetings.id = ?", meeting.ID))
	if err != nil {
		t.Error(err)
	}
	if state.Meeting.Running != false {
		t.Error("unexpected meeting state:", state)
	}
}

func TestMeetingAddAttendee(t *testing.T) {
	api, _ := NewTestRequest().Context()
	defer api.Release()

	backend := createTestBackend(api)
	meeting := createTestMeeting(api, backend)

	attendee := &bbb.Attendee{
		UserID:         "user23",
		InternalUserID: "uuu-sss-eee-rrr",
		FullName:       "User",
		Role:           "admin",
		IsPresenter:    true,
	}

	// Make request
	rpc := RPCMeetingAddAttendee(&MeetingAddAttendeeRequest{
		InternalMeetingID: meeting.InternalID,
		Attendee:          attendee,
	})
	testRPCRequest(t, rpc)

	// Get Meeting and check if attendee is in list
	tx, err := api.Conn.Begin(api.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(api.Ctx())
	state, err := store.GetMeetingState(api.Ctx(), tx, store.Q().
		Where("meetings.id = ?", meeting.ID))
	if err != nil {
		t.Error(err)
	}
	a := state.Meeting.Attendees[0]
	if a.UserID != "user23" {
		t.Error("unexpected attendee", a)
	}
	t.Log(state)
}

func TestMeetingRemoveAttendee(t *testing.T) {
	api, _ := NewTestRequest().Context()
	defer api.Release()
	ctx := api.Ctx()

	backend := createTestBackend(api)
	meeting := createTestMeeting(api, backend)

	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}

	meeting.Meeting.Attendees = append(
		meeting.Meeting.Attendees,
		&bbb.Attendee{
			UserID:         "user23",
			InternalUserID: "uuu-sss-eee-rrr",
			FullName:       "User",
		},
		&bbb.Attendee{
			UserID:         "user42",
			InternalUserID: "rrr",
			Role:           "admin",
		})

	if err := meeting.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	// Make request
	rpc := RPCMeetingRemoveAttendee(&MeetingRemoveAttendeeRequest{
		InternalMeetingID: meeting.InternalID,
		InternalUserID:    "rrr",
	})
	testRPCRequest(t, rpc)

	// Get Meeting and check if attendee is in list
	tx, err = api.Conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	state, err := store.GetMeetingState(ctx, tx, store.Q().
		Where("meetings.id = ?", meeting.ID))
	if err != nil {
		t.Error(err)
	}
	if len(state.Meeting.Attendees) != 1 {
		t.Error("unexpected attendees:", state.Meeting.Attendees)
	}
	a := state.Meeting.Attendees[0]
	if a.UserID != "user23" {
		t.Error("unexpected attendees:", state.Meeting.Attendees)
	}
	t.Log(state)
}
