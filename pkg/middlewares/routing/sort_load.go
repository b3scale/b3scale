package routing

import (
	"context"
	"sort"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// BackendsByLoad wraps a backends collection for sorting
type BackendsByLoad []*cluster.Backend

// Len returns the length of the collection
func (b BackendsByLoad) Len() int { return len(b) }

// Swap swaps two elements in the collection
func (b BackendsByLoad) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

// Less compares two backends
func (b BackendsByLoad) Less(i, j int) bool {
	return b[i].Stress() < b[j].Stress()
}

// SortLoad sorts Backends by load
func SortLoad(next cluster.RouterHandler) cluster.RouterHandler {
	return func(
		ctx context.Context,
		backends []*cluster.Backend,
		req *bbb.Request,
	) ([]*cluster.Backend, error) {
		sort.Sort(BackendsByLoad(backends))
		return next(ctx, backends, req)
	}
}
