package config

import (
	"testing"
)

func TestGetReferencedValue(t *testing.T) {
	props := Properties{
		"foo.bar": "value1",
		"key":     "${foo.bar}value2",
	}

	v1, ok := props.Get("foo.bar")
	if !ok {
		t.Error("expected value for foo.bar")
	}
	if v1 != "value1" {
		t.Error("unexpected v1 value:", v1)
	}

	v2, ok := props.Get("key")
	if !ok {
		t.Error("expected value for key")
	}
	if v2 != "value1value2" {
		t.Error("unexpected v2:", v2)
	}
}

func TestReadPropertiesFile(t *testing.T) {
	propfile := "../../testdata/config/bigbluebutton.properties"
	props, err := ReadPropertiesFile(propfile)
	if err != nil {
		t.Error(err)
	}

	v, ok := props.Get("securitySalt")
	if !ok {
		t.Error("expected presence of known key `securitySalt`")
	}
	if v != "th1sI5v3rys3cr3t" {
		t.Error("unexpected value:", v)
	}
}
