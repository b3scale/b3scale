package api

import (
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/store"
)

func createTestBackend(
	api *API,
) *store.BackendState {
	ctx := api.Ctx()
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		panic(err)
	}
	agentRef := "test-agent-2000"
	b := store.InitBackendState(&store.BackendState{
		Backend: &bbb.Backend{
			Host:   "testhost",
			Secret: "testsecret",
		},
		Settings: store.BackendSettings{
			Tags: []string{"tag1"},
		},
		AgentRef: &agentRef,
	})

	if err := b.Save(ctx, tx); err != nil {
		panic(err)
	}

	if err := tx.Commit(ctx); err != nil {
		panic(err)
	}
	return b
}

func TestBackendsList(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("admin42", auth.ScopeAdmin).
		Context()
	defer api.Release()

	// Create a backend
	createTestBackend(api)

	// List all backends
	if err := api.Handle(ResourceBackends.List); err != nil {
		t.Fatal(err)
	}

	if err := res.StatusOK(); err != nil {
		t.Fatal(err)
	}

	t.Log(res.Body())
}

func TestBackendCreate(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("admin42", auth.ScopeAdmin).
		JSON(map[string]interface{}{
			"bbb": map[string]interface{}{
				"host":   "http://testhost",
				"secret": "testsec",
			},
			"load_factor": 2.25,
			"settings":    map[string]interface{}{},
		}).
		Context()
	defer api.Release()

	if err := api.Handle(ResourceBackends.Create); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Fatal(err)
	}

	t.Log("create:", res.Body())
}

func TestBackendAgentCreate(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("context-defer-apiresource", auth.ScopeNode).
		JSON(map[string]interface{}{
			"bbb": map[string]interface{}{
				"host":   "http://testhost",
				"secret": "testsec",
			},
			"load_factor": 1.23,
		}).
		Context()
	defer api.Release()

	if err := api.Handle(ResourceBackends.Create); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Fatal(err)
	}

	body := res.JSON()
	if body["agent_ref"].(string) != "context-defer-apiresource" {
		t.Error("unexpected agent_ref", body["agent_ref"])
	}
}

func TestBackendUpdate(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("admin42", auth.ScopeAdmin).
		JSON(map[string]interface{}{
			"bbb": map[string]interface{}{
				"host": "http://newhost",
			},
		}).
		Context()
	defer api.Release()

	// Create a backend
	b := createTestBackend(api)
	t.Log("update backend id:", b.ID)

	// Create backend request
	api.SetParamNames("id")
	api.SetParamValues(b.ID)

	if err := api.Handle(ResourceBackends.Update); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Fatal(err)
	}
	t.Log("create", res.Body())
}

func TestBackendDestroy(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("admin42", auth.ScopeAdmin).
		Context()
	defer api.Release()

	// Create a backend
	b := createTestBackend(api)
	t.Log("destroy backend id:", b.ID)

	api.SetParamNames("id")
	api.SetParamValues(b.ID)

	if err := api.Handle(ResourceBackends.Destroy); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}
	t.Log("destroy:", res.Body())
}

func TestBackendForceDestroy(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("admin42", auth.ScopeAdmin).
		Query("force=true").
		Context()
	defer api.Release()

	// Create a backend
	b := createTestBackend(api)
	t.Log("force destroy backend id:", b.ID)

	api.SetParamNames("id")
	api.SetParamValues(b.ID)

	if err := api.Handle(ResourceBackends.Destroy); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}
	t.Log("destroy:", res.Body())
}

func TestBackendRetrieve(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("admin42", auth.ScopeAdmin).
		Context()
	defer api.Release()

	// Create a backend
	b := createTestBackend(api)
	t.Log("fetch backend id:", b.ID)

	// Create backend request
	api.SetParamNames("id")
	api.SetParamValues(b.ID)

	if err := api.Handle(ResourceBackends.Show); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}
	t.Log("retrieve:", res.Body())
}
