package client

import (
	"context"

	"github.com/b3scale/b3scale/pkg/http/api"
)

// Status retrievs the API / server status
func (c *Client) Status(
	ctx context.Context,
) (*api.StatusResponse, error) {
	res, err := c.Request(ctx, Fetch(""))
	if err != nil {
		return nil, err
	}
	status := &api.StatusResponse{}
	if err := res.JSON(status); err != nil {
		return nil, err
	}
	return status, nil
}
