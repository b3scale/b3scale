package main

import (
	"testing"
)

func TestParseSetProp(t *testing.T) {
	k, v := parseSetProp("foo=bar")
	if k != "foo" {
		t.Error("unexpected:", k)
	}
	if v.(string) != "bar" {
		t.Error("unexpected:", v)
	}

	k, v = parseSetProp(`foo=""`)
	if k != "foo" {
		t.Error("unexpected:", k)
	}
	if v.(string) != "" {
		t.Error("unexpected:", v)
	}

	k, v = parseSetProp("foo=")
	if k != "foo" {
		t.Error("unexpected:", k)
	}
	if v != nil {
		t.Error("unexpected:", v)
	}

	k, v = parseSetProp(`foo=["bar"]`)
	if k != "foo" {
		t.Error("unexpected:", k)
	}
	if v.([]interface{})[0].(string) != "bar" {
		t.Error("unexpected:", v)
	}

	k, v = parseSetProp(`foo={"a": 42}`)
	if k != "foo" {
		t.Error("unexpected:", k)
	}
	if v.(map[string]interface{})["a"].(float64) != 42 {
		t.Error("unexpected:", v)
	}

}
