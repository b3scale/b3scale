package client

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/b3scale/b3scale/pkg/store"
)

// Frontends creates a frontend resource
func Frontends(id ...string) string {
	return Resource("frontends", id)
}

// FrontendsList retrievs a list of frontends
func (c *Client) FrontendsList(
	ctx context.Context,
	query ...url.Values,
) ([]*store.FrontendState, error) {
	res, err := c.Request(ctx, Fetch(Frontends(), query...))
	if err != nil {
		return nil, err
	}
	frontends := []*store.FrontendState{}
	if err := res.JSON(frontends); err != nil {
		return nil, err
	}
	return frontends, nil
}

// FrontendRetrieve retrieves a single frontend
func (c *Client) FrontendRetrieve(
	ctx context.Context,
	id string,
) (*store.FrontendState, error) {
	res, err := c.Request(ctx, Fetch(Frontends(id)))
	if err != nil {
		return nil, err
	}
	frontend := &store.FrontendState{}
	if err := res.JSON(frontend); err != nil {
		return nil, err
	}
	return frontend, nil
}

// FrontendCreate POSTs a new frontend to the server
func (c *Client) FrontendCreate(
	ctx context.Context, frontend *store.FrontendState,
) (*store.FrontendState, error) {
	payload, err := json.Marshal(frontend)
	if err != nil {
		return nil, err
	}

	res, err := c.Request(ctx, Create(Frontends(), payload))
	if err != nil {
		return nil, err
	}

	frontend = &store.FrontendState{}
	if err := res.JSON(frontend); err != nil {
		return nil, err
	}
	return frontend, nil
}

// FrontendUpdateRaw PATCHes an already existing frontend
// identified by ID using raw payload.
func (c *Client) FrontendUpdateRaw(
	ctx context.Context,
	id string,
	payload []byte,
) (*store.FrontendState, error) {
	res, err := c.Request(ctx, Update(Frontends(id), payload))
	if err != nil {
		return nil, err
	}
	frontend := &store.FrontendState{}
	if err := res.JSON(frontend); err != nil {
		return nil, err
	}
	return frontend, nil
}

// FrontendUpdate PATCHes an already existing frontend.
func (c *Client) FrontendUpdate(
	ctx context.Context,
	frontend *store.FrontendState,
) (*store.FrontendState, error) {
	payload, err := json.Marshal(frontend)
	if err != nil {
		return nil, err
	}
	return c.FrontendUpdateRaw(ctx, frontend.ID, payload)
}

// FrontendDelete removes a frontend from the cluster.
func (c *Client) FrontendDelete(
	ctx context.Context, frontend *store.FrontendState,
) (*store.FrontendState, error) {
	res, err := c.Request(ctx, Destroy(Frontends(frontend.ID)))
	if err != nil {
		return nil, err
	}
	frontend = &store.FrontendState{}
	if err := res.JSON(frontend); err != nil {
		return nil, err
	}
	return frontend, nil
}
