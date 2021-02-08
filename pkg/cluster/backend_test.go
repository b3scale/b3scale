package cluster

import (
	"testing"

	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

func TestBackendStress(t *testing.T) {
	b := &Backend{state: &store.BackendState{
		ID:             "A",
		MeetingsCount:  10,
		LoadFactor:     1,
		AttendeesCount: 12,
	}}
	if b.Stress() != 22 {
		t.Error("unexpected result for stress function:", b.Stress())
	}
	b.state.LoadFactor = 1.25
	if b.Stress() != 27.5 {
		t.Error("unexpected result for stress function:", b.Stress())
	}
}
