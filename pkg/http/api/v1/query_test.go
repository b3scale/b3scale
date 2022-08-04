package v1

import (
	"net/http"
	"net/url"
	"testing"
)

func TestBackendFromQuery(t *testing.T) {
	api, _ := NewTestRequest().Context()

	backend := createTestBackend(api)

	// This should fail with a bad request
	if _, err := backendFromRequest(api.Ctx(), api, tx); err == nil {
		t.Error("no query should be a bad request")
	}

	// We need a new request, because the echo context will
	// reuse the requests query.
	u, _ := url.Parse("http:///?backend_id=" + backend.ID)
	req := &http.Request{
		URL: u,
	}
	ctx, _ = MakeTestContext(req)
	defer ctx.Release()

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
	defer ctx.Release()

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
	defer ctx.Release()
	if lookup, err := backendFromRequest(ctx, tx); err != nil {
		t.Error(err)
	} else {
		if lookup != nil {
			t.Error("expected nil, unexpected backend:", lookup)
		}
	}
}
