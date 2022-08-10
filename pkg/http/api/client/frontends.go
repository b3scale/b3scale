package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/b3scale/b3scale/pkg/store"
)

// FrontendsList retrievs a list of frontends
func (c *Client) FrontendsList(
	ctx context.Context, query url.Values,
) ([]*store.FrontendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, c.apiURL("frontends", query), nil)
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
	frontends := []*store.FrontendState{}
	err = readJSONResponse(res, &frontends)
	return frontends, err
}

// FrontendRetrieve retrieves a single frontend
func (c *Client) FrontendRetrieve(
	ctx context.Context, id string,
) (*store.FrontendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, c.apiURL("frontends/"+id, nil), nil)
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
	frontend := &store.FrontendState{}
	err = readJSONResponse(res, frontend)
	return frontend, err
}

// FrontendCreate POSTs a new frontend to the server
func (c *Client) FrontendCreate(
	ctx context.Context, frontend *store.FrontendState,
) (*store.FrontendState, error) {
	payload, err := json.Marshal(frontend)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(payload)
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, c.apiURL("frontends", nil), body)
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
	frontend = &store.FrontendState{}
	err = readJSONResponse(res, frontend)
	return frontend, err
}

// FrontendUpdateRaw PATCHes an already existing frontend
// identified by ID using raw payload.
func (c *Client) FrontendUpdateRaw(
	ctx context.Context,
	id string,
	payload []byte,
) (*store.FrontendState, error) {
	body := bytes.NewBuffer(payload)
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPatch, c.apiURL("frontends/"+id, nil), body)
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
	frontend := &store.FrontendState{}
	err = readJSONResponse(res, frontend)
	return frontend, err
}

// FrontendUpdate PATCHes an already existing frontend.
func (c *Client) FrontendUpdate(
	ctx context.Context, frontend *store.FrontendState,
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
	req, err := http.NewRequestWithContext(
		ctx, http.MethodDelete, c.apiURL("frontends/"+frontend.ID, nil), nil)
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
	frontend = &store.FrontendState{}
	err = readJSONResponse(res, frontend)
	return frontend, err
}
