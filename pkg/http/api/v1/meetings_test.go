package v1

import (
	"net/http"
	"net/url"
	"testing"

	//	"github.com/google/uuid"

	// "gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

func CreateTestMeeting(backend *store.BackendState) (*store.MeetingState, error) {
	ctx, _ := MakeTestContext(nil)
	defer ctx.Release()
	cctx := ctx.Ctx()
	tx, err := store.ConnectionFromContext(cctx).Begin(cctx)
	if err != nil {
		return nil, err
	}

	m := store.InitMeetingState(&store.MeetingState{
		BackendID: &backend.ID,
	})

	if err := m.Save(cctx, tx); err != nil {
		return nil, err
	}

	return m, tx.Commit(cctx)
}

func TestBackendFromRequest(t *testing.T) {
	if err := ClearState(); err != nil {
		t.Fatal(err)
	}

	backend, err := CreateTestBackend()
	if err != nil {
		t.Fatal(err)
	}

	ctx, _ := MakeTestContext(nil)

	cctx := ctx.Ctx()
	tx, err := store.ConnectionFromContext(cctx).Begin(cctx)
	defer tx.Rollback(cctx)

	// This should fail with a bad request
	if _, err := backendFromRequest(ctx, tx); err == nil {
		t.Error("no query should be a bad request")
	}

	// We need a new request, because the echo context will
	// reuse the requests query.
	u, _ := url.Parse("http:///?backend_id=" + backend.ID)
	req := &http.Request{
		URL: u,
	}
	ctx, _ = MakeTestContext(req)

	if lookup, err := backendFromRequest(ctx, tx); err != nil {
		t.Error(err)
	} else {
		if lookup.ID != backend.ID {
			t.Error("unexpected backend:", lookup)
		}
	}

	// We need a new request, because the echo context will
	// reuse the requests query.
	u, _ = url.Parse("http:///?backend_host=" + backend.Backend.Host)
	req = &http.Request{
		URL: u,
	}
	ctx, _ = MakeTestContext(req)

	if lookup, err := backendFromRequest(ctx, tx); err != nil {
		t.Error(err)
	} else {
		if lookup.ID != backend.ID {
			t.Error("unexpected backend", lookup)
		}
	}

	// Unknown host should yield nil
	u, _ = url.Parse("http:///?backend_host=fooo000")
	req = &http.Request{
		URL: u,
	}
	ctx, _ = MakeTestContext(req)
	if lookup, err := backendFromRequest(ctx, tx); err != nil {
		t.Error(err)
	} else {
		if lookup != nil {
			t.Error("expected nil, unexpected backend:", lookup)
		}
	}
}

func TestBackendMeetingsList(t *testing.T) {
	if err := ClearState(); err != nil {
		t.Fatal(err)
	}
	backend, err := CreateTestBackend()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(backend)
	/*
		meeting, err := createTestMeeting(backend)
		if err != nil {
			t.Fatal(err)
		}


		t.Log(meeting)
	*/
}
