package store

import "testing"

func TestSQLSafeParam(t *testing.T) {
	s := SQLSafeParam("foo`';;--")
	if s != "foo" {
		t.Error("unexpected result:", s)
	}
}
