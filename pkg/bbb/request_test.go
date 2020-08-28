package bbb

import (
	"testing"
)

func TestParamsEncode(t *testing.T) {
	params := &Params{
		"c": "foo",
		"a": 23,
		"b": true,
	}

	expected := "a=23&b=true&c=foo"
	result := params.Encode()
	if result != expected {
		t.Error("Unexpected result:", result)
	}
}
