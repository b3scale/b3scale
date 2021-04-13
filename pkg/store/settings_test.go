package store

import (
	"testing"
)

func TestTagsEq(t *testing.T) {
	t1 := Tags{"foo", "bar"}
	t2 := Tags{"bar", "foo"}

	if !t1.Eq(t2) {
		t.Error("exptected t1 == t2")
	}

	t2 = Tags{"baz"}

	if t1.Eq(t2) {
		t.Error("exptected t1 != t2")
	}
}
