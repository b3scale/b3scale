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

func TestRewriteBodyURLs(t *testing.T) {
	body := []byte(`<b><a href="/foo"></a></b>`)
	rewrite := rewriteBodyURLs(body, "b4ck3nd1d")
	if string(rewrite) != `<b><a href="/client/b4ck3nd1d/foo"></a></b>` {
		t.Error("Unexpected body:", string(rewrite))
	}
}
