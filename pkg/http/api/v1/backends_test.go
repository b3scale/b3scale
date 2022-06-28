package v1

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/store"
)

func CreateTestBackend() (*store.BackendState, error) {
	ctx, _ := MakeTestContext(nil)
	defer ctx.Release()

	cctx := ctx.Ctx()
	tx, err := store.ConnectionFromContext(cctx).Begin(cctx)
	if err != nil {
		return nil, err
	}

	b := store.InitBackendState(&store.BackendState{
		Backend: &bbb.Backend{
			Host:   "testhost",
			Secret: "testsecret",
		},
		Settings: store.BackendSettings{
			Tags: []string{"tag1"},
		},
	})

	if err := b.Save(cctx, tx); err != nil {
		return nil, err
	}

	return b, tx.Commit(cctx)
}

func TestBackendsList(t *testing.T) {
	if err := ClearState(); err != nil {
		t.Fatal(err)
	}

	ctx, rec := MakeTestContext(nil)
	defer ctx.Release()

	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})

	// Create a backend
	if _, err := CreateTestBackend(); err != nil {
		t.Fatal(err)
	}

	// List backends
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
	if err := ClearState(); err != nil {
		t.Fatal(err)
	}

	// Create backend request
	body, _ := json.Marshal(map[string]interface{}{
		"bbb": map[string]interface{}{
			"host":   "http://testhost",
			"secret": "testsec",
		},
		"load_factor": 2.25,
		"settings":    map[string]interface{}{},
	})
	req, _ := http.NewRequest("POST", "http:///", bytes.NewBuffer(body))
	req.Header.Set("content-type", "application/json")

	ctx, rec := MakeTestContext(req)
	defer ctx.Release()

	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})
	if err := BackendCreate(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
	resBody, _ := ioutil.ReadAll(res.Body)
	t.Log("create:", string(resBody))
}

func TestBackendUpdate(t *testing.T) {
	if err := ClearState(); err != nil {
		t.Fatal(err)
	}

	ctx, _ := MakeTestContext(nil)
	defer ctx.Release()

	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})

	// Create a backend
	b, err := CreateTestBackend()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("update backend id:", b.ID)

	// Create backend request
	body, _ := json.Marshal(map[string]interface{}{
		"bbb": map[string]interface{}{
			"host": "http://newhost",
		},
	})
	req, _ := http.NewRequest("PATCH", "http:///", bytes.NewBuffer(body))
	req.Header.Set("content-type", "application/json")
	ctx, rec := MakeTestContext(req)
	defer ctx.Release()

	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})
	ctx.Context.SetParamNames("id")
	ctx.Context.SetParamValues(b.ID)

	if err := BackendUpdate(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
	resBody, _ := ioutil.ReadAll(res.Body)
	t.Log("create:", string(resBody))
}

func TestBackendDestroy(t *testing.T) {
	if err := ClearState(); err != nil {
		t.Fatal(err)
	}

	ctx, _ := MakeTestContext(nil)
	defer ctx.Release()

	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})

	// Create a backend
	b, err := CreateTestBackend()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("destroy backend id:", b.ID)

	req, _ := http.NewRequest("DELETE", "http:///", nil)
	ctx, rec := MakeTestContext(req)
	defer ctx.Release()

	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})
	ctx.Context.SetParamNames("id")
	ctx.Context.SetParamValues(b.ID)

	if err := BackendDestroy(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
	resBody, _ := ioutil.ReadAll(res.Body)
	t.Log("destroy:", string(resBody))
}

func TestBackendForceDestroy(t *testing.T) {
	if err := ClearState(); err != nil {
		t.Fatal(err)
	}

	ctx, _ := MakeTestContext(nil)
	defer ctx.Release()

	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})

	// Create a backend
	b, err := CreateTestBackend()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("force destroy backend id:", b.ID)

	req, _ := http.NewRequest("DELETE", "http://test?fxrce=true", nil)
	ctx, rec := MakeTestContext(req)
	defer ctx.Release()

	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})
	ctx.Context.SetParamNames("id")
	ctx.Context.SetParamValues(b.ID)

	if err := BackendDestroy(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
	resBody, _ := ioutil.ReadAll(res.Body)
	t.Log("destroy:", string(resBody))
}

func TestBackendRetrieve(t *testing.T) {
	if err := ClearState(); err != nil {
		t.Fatal(err)
	}

	ctx, _ := MakeTestContext(nil)
	defer ctx.Release()

	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})

	// Create a backend
	b, err := CreateTestBackend()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("update backend id:", b.ID)

	// Create backend request
	req, _ := http.NewRequest("GET", "http:///", nil)
	ctx, rec := MakeTestContext(req)
	defer ctx.Release()

	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})
	ctx.Context.SetParamNames("id")
	ctx.Context.SetParamValues(b.ID)

	if err := BackendRetrieve(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
	resBody, _ := ioutil.ReadAll(res.Body)
	t.Log("retrieve:", string(resBody))
}
