package routing

import (
	"testing"

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
