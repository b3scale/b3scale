package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/b3scale/b3scale/pkg/http/api"
	"github.com/b3scale/b3scale/pkg/store"
)

// MeetingsList retrieves all meetings. Warning:
// Some scope is required otherwise the request will fail.
func (c *Client) MeetingsList(
	ctx context.Context,
	query url.Values,
) ([]*store.MeetingState, error) {
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
	return c.MeetingsList(ctx, query)
}

// MeetingRetrieve will fetch a single meeting by ID
func (c *Client) MeetingRetrieve(
	ctx context.Context,
	id string,
) (*store.MeetingState, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, c.apiURL("meetings/"+id, nil), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(c.AuthorizeRequest(req))
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if !httpSuccess(res) {
		return nil, ErrRequestFailed(res)
	}
	meeting := &store.MeetingState{}
	if err := readJSONResponse(res, meeting); err != nil {
		return nil, err
	}
	return meeting, nil
}

// InternalMeetingID returns the internal id for
// accessing via the API.
func InternalMeetingID(id string) string {
	return api.PrefixInternalID + id
}

// MeetingUpdateRaw will perform a PATCH on a meeting
func (c *Client) MeetingUpdateRaw(
	ctx context.Context,
	id string,
	payload []byte,
) (*store.MeetingState, error) {
	body := bytes.NewBuffer(payload)
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPatch, c.apiURL("meetings/"+id, nil), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.Client.Do(c.AuthorizeRequest(req))
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, ErrRequestFailed(res)
	}
	meeting := &store.MeetingState{}
	if err := readJSONResponse(res, meeting); err != nil {
		return nil, err
	}
	return meeting, nil
}

// MeetingUpdate will invoke a patch for an entire meeting
func (c *Client) MeetingUpdate(
	ctx context.Context,
	meeting *store.MeetingState,
) (*store.MeetingState, error) {
	payload, err := json.Marshal(meeting)
	if err != nil {
		return nil, err
	}
	return c.MeetingUpdateRaw(ctx, meeting.ID, payload)
}

// MeetingDelete will remove a meeting from the store
func (c *Client) MeetingDelete(
	ctx context.Context,
	id string,
) (*store.MeetingState, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodDelete, c.apiURL("meetings/"+id, nil), nil)
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
	meeting := &store.MeetingState{}
	if err := readJSONResponse(res, meeting); err != nil {
		return nil, err
	}
	return meeting, nil
}
