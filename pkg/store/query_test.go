package store

import (
	"testing"
)

func TestFilterString(t *testing.T) {
	q := &Query{offset: 1}
	f := &Filter{
		q:    q,
		idx:  1,
		attr: "foo",
		op:   "=",
	}
	s := f.String()
	if s != "foo = $3" {
		t.Error("Unexpected:", s)
	}
	t.Log(s)
}
