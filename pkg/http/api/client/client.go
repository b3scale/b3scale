package client

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// Client implements the default client for v1
type Client struct {
	Host        string
	AccessToken string

	*http.Client
}

// New initializes the client
func New(host, token string) *Client {
	return &Client{
		Host:        host,
		AccessToken: token,
		Client:      http.DefaultClient,
	}
}

// Build the request URL by joining the API base with the
// api path and resource.
func (c *Client) apiURL(resource string, query url.Values) string {
	u := c.Host
	if !strings.HasSuffix(u, "/") {
		u += "/"
	}
	u += path.Join("api/v1", resource)
	if query != nil {
		u += "?" + query.Encode()
	}
	return u
}

// AuthorizeRequest will add a http Authorization
// header with the access token to the request
func (c *Client) AuthorizeRequest(req *http.Request) *http.Request {
	bearer := "Bearer " + c.AccessToken
	req.Header.Set("Authorization", bearer)
	return req
}

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
