package cluster

import (
	"context"
)

type requestContextKey int

// Context keys for Backends and Backend
var (
	backendsContextKey = requestContextKey(1)
	backendContextKey  = requestContextKey(2)
)

// NewRequestContext create a new context
func NewRequestContext() context.Context {
	return context.Background()
}

// ContextWithBackends creates a new context from
// the parent context with a copy of the backends.
func ContextWithBackends(
	ctx context.Context, backends []*Backend,
) context.Context {
	cBackends := make([]*Backend, len(backends))
	copy(cBackends, backends)
	return context.WithValue(ctx, backendsContextKey, cBackends)
}

// BackendsFromContext retrieves backends from a context
func BackendsFromContext(ctx context.Context) []*Backend {
	backends, ok := ctx.Value(backendsContextKey).([]*Backend)
	if !ok {
		return []*Backend{}
	}
	return backends
}

// ContextWithBackend create a new context with a backend
func ContextWithBackend(
	ctx context.Context, backend *Backend,
) context.Context {
	return context.WithValue(ctx, backendContextKey, backend)
}

// BackendFromContext retrievs a backend from a context
func BackendFromContext(ctx context.Context) *Backend {
	backend, ok := ctx.Value(backendContextKey).(*Backend)
	if !ok {
		return nil
	}
	return backend
}
