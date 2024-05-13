package cluster

import (
	"context"
)

type requestContextKey int

// Context keys for Backends and Backend
var (
	backendsContextKey = requestContextKey(1)
	backendContextKey  = requestContextKey(2)
	frontendContextKey = requestContextKey(3)
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

// BackendFromContext retrieves a backend from a context
func BackendFromContext(ctx context.Context) *Backend {
	backend, ok := ctx.Value(backendContextKey).(*Backend)
	if !ok {
		return nil
	}
	return backend
}

// ContextWithFrontend creates a context with a frontend
func ContextWithFrontend(
	ctx context.Context, frontend *Frontend,
) context.Context {
	return context.WithValue(ctx, frontendContextKey, frontend)
}

// FrontendFromContext retrieves a frontend from a context
func FrontendFromContext(ctx context.Context) *Frontend {
	frontend, ok := ctx.Value(frontendContextKey).(*Frontend)
	if !ok {
		return nil
	}
	return frontend
}
