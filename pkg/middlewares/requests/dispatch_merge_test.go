package requests

import (
	"context"
	"testing"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// Create a request handler with a predefined
// set of responses.
func makeRequestHandler(responses []bbb.Response) cluster.RequestHandler {
	resIdx := 0
	return func(ctx context.Context, req *bbb.Request) (bbb.Response, error) {
		response := responses[resIdx]
		resIdx = (resIdx + 1) % len(responses)
		return response, nil
	}
}

func TestDispatchMerge(t *testing.T) {
	req := &bbb.Request{}
	// Create meeting responses
	m1 := []*bbb.Meeting{
		&bbb.Meeting{},
		&bbb.Meeting{}}
	m2 := []*bbb.Meeting{
		&bbb.Meeting{}}
	res1 := &bbb.GetMeetingsResponse{
		XMLResponse: &bbb.XMLResponse{},
		Meetings:    m1}
	res2 := &bbb.GetMeetingsResponse{
		XMLResponse: &bbb.XMLResponse{},
		Meetings:    m2}

	// Create context with two backends
	backends := []*cluster.Backend{
		&cluster.Backend{},
		&cluster.Backend{},
	}
	ctx := cluster.ContextWithBackends(
		context.Background(), backends)

	handler := makeRequestHandler([]bbb.Response{res1, res2})

	// Dispatch / Merge
	response, err := dispatchMerge(ctx, handler, req)
	if err != nil {
		t.Error(err)
	}

	getMeetings := response.(*bbb.GetMeetingsResponse)
	if len(getMeetings.Meetings) != 3 {
		t.Error("Meetings should have been merged.")
	}
}
