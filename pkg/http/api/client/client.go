package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// Resource creates a resource identifier
func Resource(res string, id []string) string {
	if len(id) > 0 {
		return res + "/" + id[0]
	}
	return res
}

// Request is an API request
type Request struct {
	Method      string
	Resource    string
	Data        []byte
	Query       url.Values
	ContentType string
}

// Fetch will create a GET request with
// optional query params.
func Fetch(resource string, query ...url.Values) *Request {
	var q url.Values
	if len(query) > 0 {
		q = query[0]
	}
	return &Request{
		Method:   http.MethodGet,
		Resource: resource,
		Query:    q,
	}
}

// Create will make a POST request
func Create(resource string, data []byte) *Request {
	return &Request{
		Method:      http.MethodPost,
		Resource:    resource,
		Data:        data,
		ContentType: "application/json",
	}
}

// Update will make a PATCH request
func Update(resource string, data []byte) *Request {
	return &Request{
		Method:      http.MethodPatch,
		Resource:    resource,
		Data:        data,
		ContentType: "application/json",
	}
}

// Destroy will make a DELETE request
func Destroy(resource string) *Request {
	return &Request{
		Method:   http.MethodDelete,
		Resource: resource,
	}
}

// Response is a http resposne
type Response http.Response

// JSON reads the entire body as json
func (res *Response) JSON(o interface{}) error {
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, o)
}

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

// Request creates a new request with context to an API url
func (c *Client) Request(
	ctx context.Context,
	req *Request,
) (*Response, error) {
	var body io.Reader
	if req.Data != nil {
		body = bytes.NewBuffer(req.Data)
	}
	// Build HTTP request with context
	httpReq, err := http.NewRequestWithContext(
		ctx,
		req.Method,
		c.apiURL(req.Resource, req.Query),
		body,
	)
	if err != nil {
		return nil, err
	}
	if req.ContentType != "" {
		httpReq.Header.Set("Content-Type", req.ContentType)
	}

	// Make request
	res, err := c.Client.Do(c.AuthorizeRequest(httpReq))
	if err != nil {
		return nil, err
	}

	// Did we 404?
	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	// Check status code for other errors
	status := res.StatusCode
	if status < 200 && status >= 400 {
		return nil, ErrRequestFailed(res)
	}

	return (*Response)(res), nil
}
