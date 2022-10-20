package schema

import "testing"

func TestGetMigrations(t *testing.T) {
	m := GetMigrations()
	if m[0].Seq != 1 {
		t.Error("unexpected migration:", m[0])
	}
	t.Log(m[0].Name)
	if m[1].Seq != 2 {
		t.Error("unexpected migration:", m[1])
	}
	t.Log(m[1].Name)
}
