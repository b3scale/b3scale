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
	if b.Stress() != 25 {
		t.Error("unexpected result for stress function:", b.Stress())
	}
	b.state.LoadFactor = 1.25
	if b.Stress() != 31.25 {
		t.Error("unexpected result for stress function:", b.Stress())
	}
}

func TestBackendHasTag(t *testing.T) {
	be := &Backend{
		state: &store.BackendState{
			Tags: []string{"foo", "bar"},
		},
	}

	if !be.HasTag("foo") {
		t.Error("backend should have tag foo")
	}

	if be.HasTag("baz") {
		t.Error("backend should not have tag baz")
	}

	be = &Backend{
		state: &store.BackendState{
			Tags: nil,
		},
	}

	if be.HasTag("foo") {
		t.Error("backend should not have tag foo")
	}
}

func TestBackendHasTags(t *testing.T) {
	be := &Backend{
		state: &store.BackendState{
			Tags: []string{"foo", "bar", "baz"},
		},
	}

	if !be.HasTags(nil) {
		t.Error("any tag might be present")
	}

	if !be.HasTags([]string{"foo", "baz"}) {
		t.Error("should have tags foo, baz")
	}

	if be.HasTags([]string{"brrz"}) {
		t.Error("should not have tags brrz")
	}

	be = &Backend{
		state: &store.BackendState{
			Tags: nil,
		},
	}

	if !be.HasTags(nil) {
		t.Error("any tag might be present")
	}

	if be.HasTags([]string{"foo"}) {
		t.Error("should not have tags foo")
	}
}
