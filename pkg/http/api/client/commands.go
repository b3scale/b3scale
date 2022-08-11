package client

import (
	"context"
	"encoding/json"

	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/store"
)

// Commands creates a new commands resource
func Commands(id ...string) string {
	return Resource("commands", id)
}

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
	res, err := c.Request(ctx, Create(Commands(), payload))
	if err != nil {
		return nil, err
	}
	cmd = &store.Command{}
	if err := res.JSON(cmd); err != nil {
		return nil, err
	}
	return cmd, nil
}

// CommandRetrieve gets a single command by ID.
// Usefull for state polling.
func (c *Client) CommandRetrieve(
	ctx context.Context,
	id string,
) (*store.Command, error) {
	res, err := c.Request(ctx, Fetch(Commands(id)))
	if err != nil {
		return nil, err
	}
	cmd := &store.Command{}
	if err := res.JSON(cmd); err != nil {
		return nil, err
	}
	return cmd, nil
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
