package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/store"
)

// CommandCreate enqueues a command into the cluster's
// command queue.
func (c *Client) CommandCreate(
	ctx context.Context,
	cmd *store.Command,
) (*store.Command, error) {
	// Encode Command
	payload, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	// Create request
	body := bytes.NewBuffer(payload)
	req, err := http.NewRequestWithContext(
		ctx, "POST", c.apiURL("commands", nil), body)
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
	cmd = &store.Command{}
	if err := readJSONResponse(res, cmd); err != nil {
		return nil, err
	}
	return cmd, err
}

// CommandRetrieve gets a single command by ID.
// Usefull for state polling.
func (c *Client) CommandRetrieve(
	ctx context.Context,
	id string,
) (*store.Command, error) {
	url := c.apiURL("commands/"+id, nil)
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
	cmd := &store.Command{}
	if err := readJSONResponse(res, cmd); err != nil {
		return nil, err
	}
	return cmd, err
}

// BackendMeetingsEnd ends all meetings on a given backend
func (c *Client) BackendMeetingsEnd(
	ctx context.Context,
	backendID string,
) (*store.Command, error) {
	cmd := cluster.EndAllMeetings(&cluster.EndAllMeetingsRequest{
		BackendID: backendID,
	})
	return c.CommandCreate(ctx, cmd)
}
