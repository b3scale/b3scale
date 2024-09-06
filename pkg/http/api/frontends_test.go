package api

import (
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/store"
)

func createTestFrontend(api *API) *store.FrontendState {
	ctx := api.Ctx()
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback(ctx)

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

	if err := f.Save(ctx, tx); err != nil {
		panic(err)
	}
	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}
	return f
}

func TestFrontendsList(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("user42", auth.ScopeAdmin).
		Context()
	defer api.Release()

	createTestFrontend(api)

	if err := api.Handle(ResourceFrontends.List); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}
	t.Log("list:", res.Body())
}

func TestFrontendsRetrieve(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("user42", auth.ScopeAdmin).
		Context()
	defer api.Release()

	f := createTestFrontend(api)

	api.SetParamNames("id")
	api.SetParamValues(f.ID)

	if err := api.Handle(ResourceFrontends.Show); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}
	t.Log("show:", res.Body())
}

func TestFrontendCreateAdmin(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("user42", auth.ScopeAdmin).
		JSON(map[string]interface{}{
			"bbb": map[string]interface{}{
				"key":    "newfrontendkey",
				"secret": "testsec",
			},
			"account_ref": "user:32421",
		}).
		Context()
	defer api.Release()

	if err := api.Handle(ResourceFrontends.Create); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}
	t.Log("create:", res.Body())
}

func TestFrontendUpdateAdmin(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("admin23", auth.ScopeAdmin).
		JSON(map[string]interface{}{
			"bbb": map[string]interface{}{
				"key":    "newkey23",
				"secret": "changedsecret",
			},
			"account_ref": "new_user_ref",
		}).
		Context()
	defer api.Release()

	// Create frontend
	f := createTestFrontend(api)

	api.SetParamNames("id")
	api.SetParamValues(f.ID)

	if err := api.Handle(ResourceFrontends.Update); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}

	data := res.JSON()
	if data["account_ref"] != "new_user_ref" {
		t.Error("unexpected account_ref", data["account_ref"])
	}
}

func TestFrontendUpdateUser(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("user23", auth.ScopeUser).
		JSON(map[string]interface{}{
			"bbb": map[string]interface{}{
				"key":    "newkey23",
				"secret": "changedsecret",
			},
			"active":      false,
			"account_ref": "new_user_ref",
		}).
		Context()
	defer api.Release()

	f := createTestFrontend(api)
	api.SetParamNames("id")
	api.SetParamValues(f.ID)

	if err := api.Handle(ResourceFrontends.Update); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}

	data := res.JSON()
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
	api, res := NewTestRequest().
		Authorize("admin42", auth.ScopeAdmin).
		Context()
	defer api.Release()

	f := createTestFrontend(api)

	api.SetParamNames("id")
	api.SetParamValues(f.ID)

	if err := api.Handle(ResourceFrontends.Destroy); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}
	t.Log("destroy:", res.Body())
}
