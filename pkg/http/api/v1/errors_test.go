package v1

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"

	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

func TestAPIErrorHandler(t *testing.T) {
	ctx, rec := MakeTestContext(nil)
	errFunc := func(_ echo.Context) error {
		return store.ValidationError{
			"fieldname": []string{
				"required", "may not be foo",
			},
		}
	}
	h := APIErrorHandler(errFunc)

	err := h(ctx)
	if err != nil {
		t.Error(err) // Error was not handled
	}

	res := rec.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Error("unexpected status code:", res.StatusCode)
	}

	body, _ := ioutil.ReadAll(res.Body)
	t.Log(string(body))
}
