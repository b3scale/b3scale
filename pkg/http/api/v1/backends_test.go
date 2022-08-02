package v1

import (
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/store"
)

func createTestBackend(
	api *APIContext,
) (*store.BackendState, error) {
	ctx := api.Ctx()
	tx, err := api.Conn.Begin(ctx)
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

	if err := b.Save(ctx, tx); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return b, nil
}

func TestBackendsList(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("admin42", ScopeAdmin).
		Context()
	defer api.Release()

	// Create a backend
	if _, err := createTestBackend(api); err != nil {
		t.Fatal(err)
	}

	// List all backends
	if err := api.Handle(APIResourceBackends.List); err != nil {
		t.Fatal(err)
	}

	if err := res.StatusOK(); err != nil {
		t.Fatal(err)
	}

	t.Log(res.Body())
}

func TestBackendCreate(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("admin42", ScopeAdmin).
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

	if err := api.Handle(APIResourceBackends.Create); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Fatal(err)
	}

	t.Log("create:", res.Body())
}

func TestBackendUpdate(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("admin42", ScopeAdmin).
		JSON(map[string]interface{}{
			"bbb": map[string]interface{}{
				"host": "http://newhost",
			},
		}).
		Context()
	defer api.Release()

	// Create a backend
	b, err := createTestBackend(api)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("update backend id:", b.ID)

	// Create backend request
	api.SetParamNames("id")
	api.SetParamValues(b.ID)

	if err := api.Handle(APIResourceBackends.Update); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Fatal(err)
	}
	t.Log("create", res.Body())
}

func TestBackendDestroy(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("admin42", ScopeAdmin).
		Context()
	defer api.Release()

	// Create a backend
	b, err := createTestBackend(api)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("destroy backend id:", b.ID)

	api.SetParamNames("id")
	api.SetParamValues(b.ID)

	if err := api.Handle(APIResourceBackends.Destroy); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}
	t.Log("destroy:", res.Body())
}

func TestBackendForceDestroy(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("admin42", ScopeAdmin).
		Query("force=true").
		Context()
	defer api.Release()

	// Create a backend
	b, err := createTestBackend(api)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("force destroy backend id:", b.ID)

	api.SetParamNames("id")
	api.SetParamValues(b.ID)

	if err := api.Handle(APIResourceBackends.Destroy); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}
	t.Log("destroy:", res.Body())
}

func TestBackendRetrieve(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("admin42", ScopeAdmin).
		Context()
	defer api.Release()

	// Create a backend
	b, err := createTestBackend(api)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("fetch backend id:", b.ID)

	// Create backend request
	api.SetParamNames("id")
	api.SetParamValues(b.ID)

	if err := api.Handle(APIResourceBackends.Show); err != nil {
		t.Fatal(err)
	}
	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}
	t.Log("retrieve:", res.Body())
}
