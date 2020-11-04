package http

import (
	"testing"
)

func TestDecodePath(t *testing.T) {
	path := "/greenlight-9b13981ff0a/bigbluebutton/api/getRecordings"
	key, action := decodePath(path)
	if key != "greenlight-9b13981ff0a" {
		t.Error("unexpected key:", key)
	}
	if action != "getRecordings" {
		t.Error("unexepcted action:", action)
	}
}
