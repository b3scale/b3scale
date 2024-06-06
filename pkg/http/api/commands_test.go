package api

import (
	"testing"

	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/http/auth"
)

func TestQueueBackendMeetingsEnd(t *testing.T) {
	cmd := cluster.EndAllMeetings(&cluster.EndAllMeetingsRequest{
		BackendID: "some-backend-id",
	})

	api, res := NewTestRequest().
		Authorize("admin42", auth.ScopeAdmin).
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
