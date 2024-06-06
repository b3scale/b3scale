package api

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/store"
)

func createTestMeeting(
	api *API,
	backend *store.BackendState,
) *store.MeetingState {
	ctx := api.Ctx()
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback(ctx)

	m := store.InitMeetingState(&store.MeetingState{
		BackendID: &backend.ID,
		Meeting: &bbb.Meeting{
			MeetingID:         uuid.New().String(),
			InternalMeetingID: uuid.New().String(),
			Running:           true,
			AttendeePW:        "foo42",
			DialNumber:        "+12 345 666",
		},
	})

	if err := m.Save(ctx, tx); err != nil {
		panic(err)
	}
	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}
	return m
}

func TestBackendMeetingsList(t *testing.T) {
	api, _ := NewTestRequest().Context()
	defer api.Release()

	backend := createTestBackend(api)
	meeting := createTestMeeting(api, backend)

	api, res := NewTestRequest().
		KeepState().
		Query("backend_host="+backend.Backend.Host).
		Authorize("admin42", auth.ScopeAdmin).
		Context()
	defer api.Release()

	if err := api.Handle(ResourceMeetings.List); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}
	body := res.Body()
	if !strings.Contains(body, meeting.ID) {
		t.Error("meeting ID", meeting.ID, "not found in response body", body)
	}
}

func TestMeetingShow(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("test-agent-2000", auth.ScopeNode).
		Context()
	defer api.Release()

	backend := createTestBackend(api)
	meeting := createTestMeeting(api, backend)

	api.SetParamNames("id")
	api.SetParamValues(InternalMeetingID(meeting.Meeting.InternalMeetingID))

	if err := api.Handle(ResourceMeetings.Show); err != nil {
		t.Fatal(err)
	}

	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}

	body := res.JSON()
	t.Log(body)
	meetingRes := body["meeting"].(map[string]interface{})
	if meetingRes["AttendeePW"].(string) != "foo42" {
		t.Error("unexpected meeting:", body)
	}
}

func TestMeetingDestroy(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("test-agent-2000", auth.ScopeNode).
		Context()
	defer api.Release()

	backend := createTestBackend(api)
	meeting := createTestMeeting(api, backend)

	api.SetParamNames("id")
	api.SetParamValues(meeting.Meeting.MeetingID)

	if err := api.Handle(ResourceMeetings.Destroy); err != nil {
		t.Fatal(err)
	}

	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}

	// Query the meeting again, this should fail.
	api, res = NewTestRequest().
		Authorize("test-agent-2000", auth.ScopeNode).
		KeepState().
		Context()
	defer api.Release()

	api.SetParamNames("id")
	api.SetParamValues(meeting.Meeting.MeetingID)

	if err := api.Handle(ResourceMeetings.Show); err == nil {
		t.Error("should raise an error")
	}
}

func TestMeetingUpdate(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("test-agent-2000", auth.ScopeNode).
		JSON(map[string]interface{}{
			"meeting": map[string]interface{}{
				"Attendees": []map[string]interface{}{
					{
						"UserID":         "user123",
						"InternalUserID": "internal-user-123",
						"FullName":       "Jen Test",
						"Role":           "admin",
						"IsPresenter":    true,
					},
					{
						"UserId":         "user42",
						"InternalUserID": "internal-user-42",
						"FullName":       "Kate Test",
						"Role":           "user",
						"IsPresenter":    false,
					},
				},
			},
		}).
		Context()
	defer api.Release()

	backend := createTestBackend(api)
	meeting := createTestMeeting(api, backend)

	api.SetParamNames("id")
	api.SetParamValues(meeting.Meeting.MeetingID)

	if err := api.Handle(ResourceMeetings.Update); err != nil {
		t.Fatal(err)
	}

	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}

	body := res.JSON()
	t.Log(body)
	meetingRes := body["meeting"].(map[string]interface{})
	attendeesRes := meetingRes["Attendees"].([]interface{})
	if len(attendeesRes) != 2 {
		t.Error("unexpected attendees", attendeesRes)
	}
	if meetingRes["DialNumber"].(string) != "+12 345 666" {
		t.Error("partial update should not have touched other props", body)
	}

}
