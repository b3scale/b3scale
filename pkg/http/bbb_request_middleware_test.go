package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/labstack/echo/v4"
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

func TestStripContentLengthHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	res := &bbb.GetMeetingsResponse{XMLResponse: &bbb.XMLResponse{}}
	header := http.Header{}

	// Set headers, testing with inconsistent case is intentional.
	header.Add("Content-Type", "application/xml")
	header.Add("Content-Length", "913")
	header.Add("content-length", "813")
	header.Add("X-BBB-Node-Test", "test")
	res.SetHeader(header)

	if err := writeBBBResponse(c, res); err != nil {
		t.Fatal(err)
	}

	r := rec.Result()
	if r.Header.Get("content-type") != "application/xml" {
		t.Error("Expected content type header to be application xml")
	}

	if r.Header.Get("Content-Length") != "" {
		t.Error("Did not expect content length in result headers")
	}

	if r.Header.Get("X-BBB-Node-Test") != "test" {
		t.Error("Expected custom header to be included")
	}
}
