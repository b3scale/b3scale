package cluster

import (
	"testing"
)

func TestContextWithBackends(t *testing.T) {
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
