package auth

import "testing"

func TestEncodeDecodeResource(t *testing.T) {
	res := EncodeResource("video", "FOO2342")
	id, format := MustDecodeResource(res)

	if id != "FOO2342" {
		t.Errorf("Expected id to be FOO2342, got %s", id)
	}
	if format != "video" {
		t.Errorf("Expected format to be video, got %s", format)
	}
}
