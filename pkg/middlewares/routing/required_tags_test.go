package routing

import (
	"testing"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

func TestRequiredTagsFromSettings(t *testing.T) {
	s := store.Settings{
		"required_tags": []string{"foo", "bar"},
	}

	tags := requiredTagsFromSettings(s)
	if tags[0] != "foo" {
		t.Error("unexpected tags", tags)
	}
}

func TestFilterRequiredTags(t *testing.T) {
	b1 := &cluster.Backend{}
	backends := []*cluster.Backend{b1}
	filtered := filterRequiredTags(backends, []string{})
	if filtered[0] != b1 {
		t.Error("unexpected:", filtered)
	}
}
