package v1

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func TestBackendsList(t *testing.T) {
	ctx, rec := MakeTestContext(nil)
	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})
	if err := BackendsList(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
	body, _ := ioutil.ReadAll(res.Body)
	t.Log(string(body))
}

func TestBackendCreate(t *testing.T) {

}
