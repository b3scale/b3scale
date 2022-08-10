package api

import (
	"testing"

	"github.com/b3scale/b3scale/pkg/cluster"
)

func TestQueueBackendMeetingsEnd(t *testing.T) {
	cmd := cluster.EndAllMeetings(&cluster.EndAllMeetingsRequest{
		BackendID: "some-backend-id",
	})

	api, res := NewTestRequest().
		Authorize("admin42", ScopeAdmin).
		JSON(cmd).
		Context()

	if err := api.Handle(ResourceCommands.Create); err != nil {
		t.Fatal(err)
	}

	if err := res.StatusOK(); err != nil {
		t.Error(err)
	}
	t.Log(res.Body())
}
