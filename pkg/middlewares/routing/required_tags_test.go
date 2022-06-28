package routing

import (
	"testing"

	"github.com/b3scale/b3scale/pkg/cluster"
)

func TestFilterRequiredTags(t *testing.T) {
	b1 := &cluster.Backend{}
	backends := []*cluster.Backend{b1}
	filtered := filterRequiredTags(backends, []string{})
	if filtered[0] != b1 {
		t.Error("unexpected:", filtered)
	}
}
