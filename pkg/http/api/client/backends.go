package client

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/b3scale/b3scale/pkg/store"
)

// Backends creates a backend resource string
func Backends(id ...string) string {
	return Resource("backends", id)
}

// BackendsList retrieves a list of backends from the server
func (c *Client) BackendsList(
	ctx context.Context,
	query ...url.Values,
) ([]*store.BackendState, error) {
	res, err := c.Request(ctx, Fetch(Backends(), query...))
	if err != nil {
		return nil, err
	}
	backends := []*store.BackendState{}
	if err := res.JSON(&backends); err != nil {
		return nil, err
	}
	return backends, nil
}

// BackendRetrieve retrieves a single backend by ID.
func (c *Client) BackendRetrieve(
	ctx context.Context,
	id string,
) (*store.BackendState, error) {
	res, err := c.Request(ctx, Fetch(Backends(id)))
	if err != nil {
		return nil, err
	}
	backend := &store.BackendState{}
	if err := res.JSON(backend); err != nil {
		return nil, err
	}
	return backend, nil
}

// BackendCreate creates a new backend on the server
func (c *Client) BackendCreate(
	ctx context.Context,
	backend *store.BackendState,
) (*store.BackendState, error) {
	payload, err := json.Marshal(backend)
	if err != nil {
		return nil, err
	}
	res, err := c.Request(ctx, Create(Backends(), payload))
	if err != nil {
		return nil, err
	}
	backend = &store.BackendState{}
	if err := res.JSON(backend); err != nil {
		return nil, err
	}
	return backend, nil
}

// BackendUpdateRaw updates an existing backend
// identified by ID with a raw JSON payload.
func (c *Client) BackendUpdateRaw(
	ctx context.Context,
	id string,
	payload []byte,
) (*store.BackendState, error) {
	res, err := c.Request(ctx, Update(Backends(id), payload))
	if err != nil {
		return nil, err
	}
	backend := &store.BackendState{}
	if err := res.JSON(backend); err != nil {
		return nil, err
	}
	return backend, nil
}

// BackendUpdate updates the backend
func (c *Client) BackendUpdate(
	ctx context.Context, backend *store.BackendState,
) (*store.BackendState, error) {
	payload, err := json.Marshal(backend)
	if err != nil {
		return nil, err
	}
	return c.BackendUpdateRaw(ctx, backend.ID, payload)
}

// BackendDelete removes a backend from the cluster
func (c *Client) BackendDelete(
	ctx context.Context,
	backend *store.BackendState,
	opts ...url.Values,
) (*store.BackendState, error) {
	res, err := c.Request(ctx, Destroy(Backends(backend.ID), opts...))
	if err != nil {
		return nil, err
	}
	backend = &store.BackendState{}
	if err := res.JSON(backend); err != nil {
		return nil, err
	}
	return backend, nil
}
