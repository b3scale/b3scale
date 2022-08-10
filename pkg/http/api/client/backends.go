package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/b3scale/b3scale/pkg/store"
)

// BackendsList retrievs a list of backends from the server
func (c *Client) BackendsList(
	ctx context.Context, query url.Values,
) ([]*store.BackendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, c.apiURL("backends", query), nil)
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
	backends := []*store.BackendState{}
	err = readJSONResponse(res, &backends)
	return backends, err
}

// BackendRetrieve retrieves a single backend by ID.
func (c *Client) BackendRetrieve(
	ctx context.Context, id string,
) (*store.BackendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, c.apiURL("backends/"+id, nil), nil)
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
	backend := &store.BackendState{}
	err = readJSONResponse(res, backend)
	return backend, err
}

// BackendCreate creates a new backend on the server
func (c *Client) BackendCreate(
	ctx context.Context, backend *store.BackendState,
) (*store.BackendState, error) {
	payload, err := json.Marshal(backend)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(payload)
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, c.apiURL("backends", nil), body)
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
	backend = &store.BackendState{}
	err = readJSONResponse(res, backend)
	return backend, err
}

// BackendUpdateRaw updates an existing backend
// identified by ID with a raw JSON payload.
func (c *Client) BackendUpdateRaw(
	ctx context.Context,
	id string,
	payload []byte,
) (*store.BackendState, error) {
	body := bytes.NewBuffer(payload)
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPatch, c.apiURL("backends/"+id, nil), body)
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
	backend := &store.BackendState{}
	err = readJSONResponse(res, backend)
	return backend, err
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
	query url.Values,
) (*store.BackendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodDelete, c.apiURL("backends/"+backend.ID, query), nil)
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
	backend = &store.BackendState{}
	err = readJSONResponse(res, backend)
	return backend, err
}
