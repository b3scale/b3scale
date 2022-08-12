package client

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/b3scale/b3scale/pkg/http/api"
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

// AgentRPC makes an rpc call
func (c *Client) AgentRPC(
	ctx context.Context,
	req *api.RPCRequest,
) (api.RPCResult, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	res, err := c.Request(ctx, Create("agent/rpc", payload))
	if err != nil {
		return nil, err
	}
	// Decode response
	rpc := &api.RPCResponse{}
	if err := res.JSON(rpc); err != nil {
		return nil, err
	}
	if rpc.Status == api.RPCStatusError {
		msg, ok := rpc.Result.(string)
		if !ok {
			msg = "unknown error"
		}
		return nil, errors.New(msg)
	}

	return rpc.Result, nil
}
