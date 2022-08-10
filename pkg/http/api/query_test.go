package api

import (
	"testing"
)

func TestBackendFromQuery(t *testing.T) {
	api, _ := NewTestRequest().Context()
	defer api.Release()

	ctx := api.Ctx()
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	backend := createTestBackend(api)

	// This should fail with a bad request
	if _, err := BackendFromQuery(ctx, api, tx); err == nil {
		t.Error("no query should be a bad request")
	}

	// New request: Query by ID
	api, _ = NewTestRequest().
		KeepState().
		Query("backend_id=" + backend.ID).
		Context()
	defer api.Release()
	ctx = api.Ctx()
	tx, err = api.Conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	if lookup, err := BackendFromQuery(ctx, api, tx); err != nil {
		t.Error(err)
	} else {
		if lookup.ID != backend.ID {
			t.Error("unexpected backend:", lookup)
		}
	}

	// New request: Query by host
	api, _ = NewTestRequest().
		KeepState().
		Query("backend_host=" + backend.Backend.Host).
		Context()
	defer api.Release()
	ctx = api.Ctx()
	tx, err = api.Conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if lookup, err := BackendFromQuery(ctx, api, tx); err != nil {
		t.Error(err)
	} else {
		if lookup.ID != backend.ID {
			t.Error("unexpected backend", lookup)
		}
	}

	// Unknown host should yield nil
	api, _ = NewTestRequest().
		KeepState().
		Query("backend_host=f000000bar").
		Context()
	defer api.Release()
	ctx = api.Ctx()
	tx, err = api.Conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if lookup, err := BackendFromQuery(ctx, api, tx); err != nil {
		t.Error(err)
	} else {
		if lookup != nil {
			t.Error("expected nil, unexpected backend:", lookup)
		}
	}
}
