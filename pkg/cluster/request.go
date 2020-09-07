package cluster

import (
	"context"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The RequestContext is a golang Context and
// can be used as such. It provides convenient methods
// for accessing values.
type RequestContext context.Context

// A Request is a request to the cluster, containing
// the BBB api request.
type Request struct {
	*bbb.Request
	Context context.Context
}

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
