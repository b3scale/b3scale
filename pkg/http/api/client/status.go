package client

import (
	"context"
	"net/http"

	"github.com/b3scale/b3scale/pkg/http/api"
)

// Status retrievs the API / server status
func (c *Client) Status(
	ctx context.Context,
) (*api.StatusResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, c.apiURL("", nil), nil)
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
	status := &api.StatusResponse{}
	err = readJSONResponse(res, status)
	return status, err
}
