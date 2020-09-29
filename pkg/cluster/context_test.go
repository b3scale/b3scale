package cluster

import (
	"testing"
)

func TestContextBackends(t *testing.T) {
	backends := []*Backend{
		&Backend{}, &Backend{},
	}

	ctx := NewRequestContext()
	ctx = ContextWithBackends(ctx, backends)

	backends1 := BackendsFromContext(ctx)
	if len(backends1) != len(backends) {
		t.Error("length should match")
	}
}

func TestContextBackend(t *testing.T) {
	backend := &Backend{}
	ctx := NewRequestContext()
	ctx = ContextWithBackend(ctx, backend)
	backend1 := BackendFromContext(ctx)
	if backend1 == nil {
		t.Error("backend should be present")
	}
}

func TestContextFrontend(t *testing.T) {
	frontend := &Frontend{}
	ctx := NewRequestContext()
	if FrontendFromContext(ctx) != nil {
		t.Error("context should not have a frontend")
	}
	ctx = ContextWithFrontend(ctx, frontend)
	if FrontendFromContext(ctx) == nil {
		t.Error("context should have a frontend")
	}
}
