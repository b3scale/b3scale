package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestUpdateCallbackQuery(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet,
		"/endpoint?foo=42&meeting=test",
		nil)

	ctx := e.NewContext(req, httptest.NewRecorder())

	cb, err := updateCallbackQuery(ctx, "/callback")
	if err != nil {
		t.Fatal(err)
	}

	// Check if params are merged
	if !strings.Contains(cb, "foo=42") {
		t.Error("unexpected url:", cb)
	}
	if !strings.Contains(cb, "meeting=test") {
		t.Error("unexpected url:", cb)
	}
	t.Log(cb)

	// Check if query params are overwritten
	cb, err = updateCallbackQuery(ctx, "/callback?meetingID=42")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cb, "meetingID=42") {
		t.Error("unexpected url:", cb)
	}

	t.Log(cb)
}
