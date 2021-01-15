package http

import (
	"testing"
)

func TestDecodeClientProxyPath(t *testing.T) {
	path := "/client/b4ck3ndID/foo/bar"
	bID, p := decodeClientProxyPath("/client", path)
	if bID != "b4ck3ndID" {
		t.Error("unexpected backendID", bID)
	}
	if p != "/foo/bar" {
		t.Error("unexpected path", p)
	}

	path = "/client/backendID"
	bID, p = decodeClientProxyPath("/client", path)
	if bID != "backendID" {
		t.Error("unexpected backendID", bID)
	}
	if p != "/" {
		t.Error("unexpected path", p)
	}

	path = "/client/id/"
	bID, p = decodeClientProxyPath("/client", path)
	if bID != "id" {
		t.Error("unexpected backendID", bID)
	}
	if p != "/" {
		t.Error("unexpected path", p)
	}
}
