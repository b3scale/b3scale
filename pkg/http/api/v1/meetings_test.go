package v1

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/store"
)

func createTestMeeting(
	api *APIContext,
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
		Authorize("admin42", ScopeAdmin).
		Context()
	defer api.Release()

	if err := api.Handle(APIResourceMeetings.List); err != nil {
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
