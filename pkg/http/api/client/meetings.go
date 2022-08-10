package client

import (
	"context"
	"net/http"
	"net/url"

	"github.com/b3scale/b3scale/pkg/store"
)

// BackendMeetingsList retrieves all meetings for a given backend
func (c *Client) BackendMeetingsList(
	ctx context.Context,
	backendID string,
	query url.Values,
) ([]*store.MeetingState, error) {
	if query == nil {
		query = url.Values{}
	}
	query.Set("backend_id", backendID)

	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, c.apiURL("meetings", query), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(c.AuthorizeRequest(req))
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, ErrRequestFailed(res)
	}
	meetings := []*store.MeetingState{}
	err = readJSONResponse(res, meetings)
	return meetings, err
}
