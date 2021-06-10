package v1

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

func createTestFrontend() (*store.FrontendState, error) {
	ctx, _ := MakeTestContext(nil)
	defer ctx.Release()
	cctx := ctx.Ctx()
	tx, err := store.ConnectionFromContext(cctx).Begin(cctx)
	if err != nil {
		return nil, err
	}

	ref := "user23"
	f := store.InitFrontendState(&store.FrontendState{
		Frontend: &bbb.Frontend{
			Key:    "testkey",
			Secret: "testsecret",
		},
		Active: true,
		Settings: store.FrontendSettings{
			RequiredTags: []string{"tag1"},
		},
		AccountRef: &ref,
	})

	if err := f.Save(cctx, tx); err != nil {
		return nil, err
	}

	return f, tx.Commit(cctx)
}

func clearFrontends() error {
	ctx, _ := MakeTestContext(nil)
	defer ctx.Release()

	reqCtx := ctx.Ctx()
	tx, err := store.ConnectionFromContext(reqCtx).Begin(reqCtx)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(reqCtx, "DELETE FROM frontends"); err != nil {
		return err
	}
	if err := tx.Commit(reqCtx); err != nil {
		return err
	}
	return nil
}

func TestFrontendsList(t *testing.T) {
	if err := clearFrontends(); err != nil {
		t.Fatal(err)
	}
	if _, err := createTestFrontend(); err != nil {
		t.Fatal(err)
	}

	ctx, rec := MakeTestContext(nil)
	ctx = AuthorizeTestContext(ctx, "user42", []string{ScopeAdmin})
	defer ctx.Release()

	if err := FrontendsList(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}

	resBody, _ := ioutil.ReadAll(res.Body)
	t.Log("list:", string(resBody))
}

func TestFrontendsRetrieve(t *testing.T) {
	if err := clearFrontends(); err != nil {
		t.Fatal(err)
	}
	f, err := createTestFrontend()
	if err != nil {
		t.Fatal(err)
	}

	ctx, rec := MakeTestContext(nil)
	ctx = AuthorizeTestContext(ctx, "user42", []string{ScopeAdmin})
	defer ctx.Release()

	ctx.Context.SetParamNames("id")
	ctx.Context.SetParamValues(f.ID)

	if err := FrontendRetrieve(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}

	resBody, _ := ioutil.ReadAll(res.Body)
	t.Log("list:", string(resBody))
}

func TestFrontendCreateAdmin(t *testing.T) {
	if err := clearFrontends(); err != nil {
		t.Fatal(err)
	}

	// Create backend request
	body, _ := json.Marshal(map[string]interface{}{
		"bbb": map[string]interface{}{
			"key":    "newfrontendkey",
			"secret": "testsec",
		},
		"account_ref": "user:32421",
	})
	req, _ := http.NewRequest("POST", "http:///", bytes.NewBuffer(body))
	req.Header.Set("content-type", "application/json")

	ctx, rec := MakeTestContext(req)
	defer ctx.Release()

	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})
	if err := FrontendCreate(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
	resBody, _ := ioutil.ReadAll(res.Body)
	t.Log("create:", string(resBody))
}

func TestFrontendCreateUser(t *testing.T) {
	if err := clearFrontends(); err != nil {
		t.Fatal(err)
	}

	// Create backend request
	body, _ := json.Marshal(map[string]interface{}{
		"bbb": map[string]interface{}{
			"key":    "newfrontendkey",
			"secret": "testsec",
		},
		"account_ref": "admin42",
	})
	req, _ := http.NewRequest("POST", "http:///", bytes.NewBuffer(body))
	req.Header.Set("content-type", "application/json")

	ctx, rec := MakeTestContext(req)
	defer ctx.Release()

	ctx = AuthorizeTestContext(ctx, "user23", []string{})
	if err := FrontendCreate(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
	resBody, _ := ioutil.ReadAll(res.Body)
	t.Log("create:", string(resBody))
	data := map[string]interface{}{}
	json.Unmarshal(resBody, &data)

	if data["account_ref"] != "user23" {
		t.Error("unexpected account_ref", data["account_ref"])
	}
}

func TestFrontendUpdateAdmin(t *testing.T) {
	if err := clearFrontends(); err != nil {
		t.Fatal(err)
	}

	f, err := createTestFrontend()
	if err != nil {
		t.Fatal(err)
	}

	// Create backend request
	body, _ := json.Marshal(map[string]interface{}{
		"bbb": map[string]interface{}{
			"key":    "newkey23",
			"secret": "changedsecret",
		},
		"account_ref": "new_user_ref",
	})
	req, _ := http.NewRequest("POST", "http:///", bytes.NewBuffer(body))
	req.Header.Set("content-type", "application/json")
	ctx, rec := MakeTestContext(req)
	defer ctx.Release()
	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})

	ctx.Context.SetParamNames("id")
	ctx.Context.SetParamValues(f.ID)

	if err := FrontendUpdate(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
	resBody, _ := ioutil.ReadAll(res.Body)
	t.Log("update:", string(resBody))
	data := map[string]interface{}{}
	json.Unmarshal(resBody, &data)

	if data["account_ref"] != "new_user_ref" {
		t.Error("unexpected account_ref", data["account_ref"])
	}
}

func TestFrontendUpdateUser(t *testing.T) {
	if err := clearFrontends(); err != nil {
		t.Fatal(err)
	}

	f, err := createTestFrontend()
	if err != nil {
		t.Fatal(err)
	}

	// Create backend request
	body, _ := json.Marshal(map[string]interface{}{
		"bbb": map[string]interface{}{
			"key":    "newkey23",
			"secret": "changedsecret",
		},
		"active":      false,
		"account_ref": "new_user_ref",
	})
	req, _ := http.NewRequest("POST", "http:///", bytes.NewBuffer(body))
	req.Header.Set("content-type", "application/json")
	ctx, rec := MakeTestContext(req)
	defer ctx.Release()
	ctx = AuthorizeTestContext(ctx, "user23", []string{})

	ctx.Context.SetParamNames("id")
	ctx.Context.SetParamValues(f.ID)

	if err := FrontendUpdate(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
	resBody, _ := ioutil.ReadAll(res.Body)
	t.Log("update:", string(resBody))
	data := map[string]interface{}{}
	json.Unmarshal(resBody, &data)

	if data["account_ref"] != "user23" {
		t.Error("unexpected account_ref", data["account_ref"])
	}
	if data["active"] != false {
		t.Error("active should be false")
	}
	if data["bbb"].(map[string]interface{})["key"] != "newkey23" {
		t.Error("unexpected bbb.key")
	}
}

func TestFrontendDestroy(t *testing.T) {
	if err := clearFrontends(); err != nil {
		t.Fatal(err)
	}

	f, err := createTestFrontend()
	if err != nil {
		t.Fatal(err)
	}

	// Create request
	req, _ := http.NewRequest("DELETE", "http:///", nil)

	ctx, rec := MakeTestContext(req)
	defer ctx.Release()
	ctx = AuthorizeTestContext(ctx, "admin42", []string{ScopeAdmin})

	ctx.Context.SetParamNames("id")
	ctx.Context.SetParamValues(f.ID)

	if err := FrontendDestroy(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
	resBody, _ := ioutil.ReadAll(res.Body)
	t.Log("destroy:", string(resBody))
}
