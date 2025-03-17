package client

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/http/api"
	"github.com/b3scale/b3scale/pkg/store"
)

// Recordings creates a recording resource URL
func Recordings(id ...string) string {
	return Resource("recordings", id)
}

// RecordingsListByFrontendID retrieves all recordings for
// a frontend identified by ID.
func (c *Client) RecordingsListByFrontendID(
	ctx context.Context,
	feID string,
) ([]*store.RecordingState, error) {
	qry := url.Values{}
	qry.Add("frontend_id", feID)
	res, err := c.Request(ctx, Fetch(Recordings(), qry))
	if err != nil {
		return nil, err
	}

	recs := []*store.RecordingState{}
	if err := res.JSON(&recs); err != nil {
		return nil, err
	}

	return recs, nil
}

// RecordingsListByFrontendKey retrieves all recordings for
// a frontend identified by ID.
func (c *Client) RecordingsListByFrontendKey(
	ctx context.Context,
	feKey string,
) ([]*store.RecordingState, error) {
	qry := url.Values{}
	qry.Add("frontend_key", feKey)
	res, err := c.Request(ctx, Fetch(Recordings(), qry))
	if err != nil {
		return nil, err
	}

	recs := []*store.RecordingState{}
	if err := res.JSON(&recs); err != nil {
		return nil, err
	}

	return recs, nil
}

// RecordingsRetrieve fetches a single recording.
func (c *Client) RecordingsRetrieve(
	ctx context.Context,
	id string,
) (*store.RecordingState, error) {
	res, err := c.Request(ctx, Fetch(Recordings(id)))
	if err != nil {
		return nil, err
	}

	rec := &store.RecordingState{}
	if err := res.JSON(rec); err != nil {
		return nil, err
	}

	return rec, nil
}

// RecordingsSetVisibility sets the visibility of a recording
func (c *Client) RecordingsSetVisibility(
	ctx context.Context,
	id string,
	v bbb.RecordingVisibility,
) (*store.RecordingState, error) {
	update := api.RecordingVisibilityUpdate{
		RecordID:   id,
		Visibility: v,
	}
	payload, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	res, err := c.Request(ctx, Create("recordings-visibility", payload))
	if err != nil {
		return nil, err
	}

	rec := &store.RecordingState{}
	if err := res.JSON(rec); err != nil {
		return nil, err
	}

	return rec, nil
}
