package client

import (
	"context"

	"github.com/b3scale/b3scale/pkg/store"
)

// AgentHeartbeatCreate creates a heartbeat for a backend
func (c *Client) AgentHeartbeatCreate(
	ctx context.Context,
) (*store.AgentHeartbeat, error) {
	res, err := c.Request(ctx, Fetch("agent/heartbeat", nil))
	if err != nil {
		return nil, err
	}
	heartbeat := &store.AgentHeartbeat{}
	if err := res.JSON(heartbeat); err != nil {
		return nil, err
	}
	return heartbeat, nil
}
