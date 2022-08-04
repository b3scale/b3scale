package v1

import (
	"io/ioutil"
	"net/http"
	"net/url"
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
	if err := ClearState(); err != nil {
		t.Fatal(err)
	}
	backend, err := CreateTestBackend()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CreateTestMeeting(backend); err != nil {
		t.Fatal(err)
	}

	u, _ := url.Parse("http:///?backend_host=" + backend.Backend.Host)
	req := &http.Request{
		URL: u,
	}
	ctx, rec := MakeTestContext(req)
	defer ctx.Release()
	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})

	if err := BackendMeetingsList(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
	resBody, _ := ioutil.ReadAll(res.Body)
	t.Log("list:", string(resBody))
}

func TestBackendMeetingsEnd(t *testing.T) {
	if err := ClearState(); err != nil {
		t.Fatal(err)
	}
	backend, err := CreateTestBackend()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CreateTestMeeting(backend); err != nil {
		t.Fatal(err)
	}

	u, _ := url.Parse("http:///?backend_host=" + backend.Backend.Host)
	req := &http.Request{
		URL: u,
	}
	ctx, rec := MakeTestContext(req)
	defer ctx.Release()
	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})

	if err := BackendMeetingsEnd(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusAccepted {
		t.Error("unexpected status code:", res.StatusCode)
	}
	resBody, _ := ioutil.ReadAll(res.Body)
	t.Log("list:", string(resBody))

}
