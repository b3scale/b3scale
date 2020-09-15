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
