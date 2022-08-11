package api

import "testing"

func TestAgentHeartbeatCreate(t *testing.T) {
	api, res := NewTestRequest().
		Authorize("test-agent-2000", ScopeNode).
		Context()
	defer api.Release()

	backend := createTestBackend(api)

	if err := api.Handle(ResourceAgentHeartbeat.Create); err != nil {
		t.Fatal(err)
	}

	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}

	body := res.JSON()
	if body["backend_id"].(string) != backend.ID {
		t.Error("unexpected backend:", body)
	}
	t.Log(body)
}
