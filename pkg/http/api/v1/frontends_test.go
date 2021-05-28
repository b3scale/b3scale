package v1

import (
	"net/http"
	"testing"
)

func TestFrontendsList(t *testing.T) {
	ctx, rec := MakeTestContext(nil)
	ctx = AuthorizeTestContext(ctx, "user42", []string{ScopeAdmin})
	if err := FrontendsList(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
}
