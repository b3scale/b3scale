package v1

import (
	"net/http"
	"net/url"
	"testing"

	//	"github.com/google/uuid"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

func createTestMeeting() (*store.FrontendState, error) {
	ctx, _ := MakeTestContext(nil)
	defer ctx.Release()
	cctx := ctx.Ctx()
	tx, err := store.ConnectionFromContext(cctx).Begin(cctx)
	if err != nil {
		return nil, err
	}

	ref := "user23"
	f := store.InitFrontendState(&store.FrontendState{
		Frontend: &bbb.Frontend{
			Key:    "testkey",
			Secret: "testsecret",
		},
		Active: true,
		Settings: store.FrontendSettings{
			RequiredTags: []string{"tag1"},
		},
		AccountRef: &ref,
	})

	if err := f.Save(cctx, tx); err != nil {
		return nil, err
	}

	return f, tx.Commit(cctx)
}

func clearMeetings() error {
	ctx, _ := MakeTestContext(nil)
	defer ctx.Release()

	reqCtx := ctx.Ctx()
	tx, err := store.ConnectionFromContext(reqCtx).Begin(reqCtx)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(reqCtx, "DELETE FROM meetings"); err != nil {
		return err
	}
	if err := tx.Commit(reqCtx); err != nil {
		return err
	}
	return nil
}

func TestBackendFromRequest(t *testing.T) {
	if err := clearBackends(); err != nil {
		t.Fatal(err)
	}

	backend, err := createTestBackend()
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
