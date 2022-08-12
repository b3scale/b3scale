package client

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/b3scale/b3scale/pkg/store"
)

// Meetings creates a meetings resource
func Meetings(id ...string) string {
	return Resource("meetings", id)
}

// MeetingsList retrieves all meetings. Warning:
// Some scope is required otherwise the request will fail.
func (c *Client) MeetingsList(
	ctx context.Context,
	query ...url.Values,
) ([]*store.MeetingState, error) {
	res, err := c.Request(ctx, Fetch(Meetings(), query...))
	if err != nil {
		return nil, err
	}
	meetings := []*store.MeetingState{}
	if err := res.JSON(&meetings); err != nil {
		return nil, err
	}
	return meetings, nil
}

// BackendMeetingsList retrieves all meetings for a given backend
func (c *Client) BackendMeetingsList(
	ctx context.Context,
	backendID string,
	query ...url.Values,
) ([]*store.MeetingState, error) {
	if len(query) == 0 {
		query = append(query, url.Values{})
	}
	query[0].Set("backend_id", backendID)
	return c.MeetingsList(ctx, query...)
}

// MeetingRetrieve will fetch a single meeting by ID
func (c *Client) MeetingRetrieve(
	ctx context.Context,
	id string,
) (*store.MeetingState, error) {
	res, err := c.Request(ctx, Fetch(Meetings(id)))
	if err != nil {
		return nil, err
	}
	meeting := &store.MeetingState{}
	if err := res.JSON(meeting); err != nil {
		return nil, err
	}
	return meeting, nil
}

// MeetingUpdateRaw will perform a PATCH on a meeting
func (c *Client) MeetingUpdateRaw(
	ctx context.Context,
	id string,
	payload []byte,
) (*store.MeetingState, error) {
	res, err := c.Request(ctx, Update(Meetings(id), payload))
	if err != nil {
		return nil, err
	}
	meeting := &store.MeetingState{}
	if err := res.JSON(meeting); err != nil {
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
	res, err := c.Request(ctx, Destroy(Meetings(id)))
	if err != nil {
		return nil, err
	}

	meeting := &store.MeetingState{}
	if err := res.JSON(meeting); err != nil {
		return nil, err
	}
	return meeting, nil
}
