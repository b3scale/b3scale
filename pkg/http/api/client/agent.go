package client

import (
	"context"

	"github.com/b3scale/b3scale/pkg/store"
)

// AgentHeartbeatCreate creates a heartbeat for a backend
func (c *Client) AgentHeartbeatCreate(
	ctx context.Context,
) (*store.AgentHeartbeat, error) {
	res, err := c.Request(ctx, Create("agent/heartbeat", nil))
	if err != nil {
		return nil, err
	}
	heartbeat := &store.AgentHeartbeat{}
	if err := res.JSON(heartbeat); err != nil {
		return nil, err
	}
	return heartbeat, nil
}

// AgentBackendRetrieve fetches the currently registered backend
func (c *Client) AgentBackendRetrieve(
	ctx context.Context,
) (*store.BackendState, error) {
	res, err := c.Request(ctx, Fetch("agent/backend"))
	if err != nil {
		return nil, err
	}
	backend := &store.BackendState{}
	if err := res.JSON(backend); err != nil {
		return nil, err
	}
	return backend, nil
}
