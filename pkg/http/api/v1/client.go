package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// Client is an interface to the v1 API.
type Client interface {
	FrontendsList(
		ctx context.Context, query url.Values,
	) ([]*store.FrontendState, error)
	FrontendRetrieve(
		ctx context.Context, id string,
	) (*store.FrontendState, error)
	FrontendCreate(
		ctx context.Context, frontend *store.FrontendState,
	) (*store.FrontendState, error)
	FrontendUpdate(
		ctx context.Context, frontend *store.FrontendState,
	) (*store.FrontendState, error)
	FrontendDelete(
		ctx context.Context, frontend *store.FrontendState,
	) (*store.FrontendState, error)

	BackendsList(
		ctx context.Context, query url.Values,
	) ([]*store.BackendState, error)
	BackendRetrieve(
		ctx context.Context, id string,
	) (*store.BackendState, error)
	BackendCreate(
		ctx context.Context, backend *store.BackendState,
	) (*store.BackendState, error)
	BackendUpdate(
		ctx context.Context, backend *store.BackendState,
	) (*store.BackendState, error)
	BackendDelete(
		ctx context.Context, backend *store.BackendState,
		query url.Values,
	) (*store.BackendState, error)

	BackendMeetingsList(
		ctx context.Context,
		backendID string,
		query url.Values,
	) ([]*store.MeetingState, error)

	BackendMeetingsEnd(
		ctx context.Context,
		backendID string,
	) (*store.Command, error)
}

// JSON helper
func readJSONResponse(res *http.Response, data interface{}) error {
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, data)
}

// Status helper
func httpSuccess(res *http.Response) bool {
	s := res.StatusCode
	return s >= 200 && s < 400
}

// JWTClient is a http api v1 client
type JWTClient struct {
	Host        string
	AccessToken string

	Client *http.Client
}

// APIError will return the decoded json body
// when the response status was not OK or Accepted
type APIError map[string]interface{}

// APIErrorFromResponse will create a new APIError from the
// HTTP response.
func APIErrorFromResponse(res *http.Response) APIError {
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return APIError{"error": err.Error()}
	}
	apiErr := APIError{}
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return APIError{"error": body}
	}
	return apiErr
}

// Error implements the error interface
func (err APIError) Error() string {
	errs := []string{}
	for k, v := range err {
		errs = append(errs, fmt.Sprintf("%s: %v", k, v))
	}
	return strings.Join(errs, "; ")
}

// NewJWTClient initializes the client
func NewJWTClient(host, token string) *JWTClient {
	return &JWTClient{
		Host:        host,
		AccessToken: token,
		Client:      http.DefaultClient,
	}
}

func (c *JWTClient) apiURL(resource string, query url.Values) string {
	u := c.Host
	if !strings.HasSuffix(u, "/") {
		u += "/"
	}
	u += "api/v1/" + resource

	if query != nil {
		u += "?" + query.Encode()
	}

	return u
}

// FrontendsList retrievs a list of frontends
func (c *JWTClient) FrontendsList(
	ctx context.Context, query url.Values,
) ([]*store.FrontendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, "GET", c.apiURL("frontends", query), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, APIErrorFromResponse(res)
	}
	frontends := []*store.FrontendState{}
	err = readJSONResponse(res, frontends)
	return frontends, err
}

// FrontendRetrieve retrieves a single frontend
func (c *JWTClient) FrontendRetrieve(
	ctx context.Context, id string,
) (*store.FrontendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, "GET", c.apiURL("frontends/"+id, nil), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, APIErrorFromResponse(res)
	}
	frontend := &store.FrontendState{}
	err = readJSONResponse(res, frontend)
	return frontend, err
}

// FrontendCreate POSTs a new frontend to the server
func (c *JWTClient) FrontendCreate(
	ctx context.Context, frontend *store.FrontendState,
) (*store.FrontendState, error) {
	return nil, nil
}

// FrontendUpdate PATCHes an already existing frontend.
func (c *JWTClient) FrontendUpdate(
	ctx context.Context, frontend *store.FrontendState,
) (*store.FrontendState, error) {
	return nil, nil
}

// FrontendDelete removes a frontend from the cluster.
func (c *JWTClient) FrontendDelete(
	ctx context.Context, frontend *store.FrontendState,
) (*store.FrontendState, error) {
	return nil, nil
}

// BackendsList retrievs a list of backends from the server
func (c *JWTClient) BackendsList(
	ctx context.Context, query url.Values,
) ([]*store.BackendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, "GET", c.apiURL("backends", query), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, APIErrorFromResponse(res)
	}
	backends := []*store.BackendState{}
	err = readJSONResponse(res, backends)
	return backends, err
}

// BackendRetrieve retrieves a single backend by ID.
func (c *JWTClient) BackendRetrieve(
	ctx context.Context, id string,
) (*store.BackendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, "GET", c.apiURL("backends/"+id, nil), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, APIErrorFromResponse(res)
	}
	backend := &store.BackendState{}
	err = readJSONResponse(res, backend)
	return backend, err
}

// BackendCreate creates a new backend on the server
func (c *JWTClient) BackendCreate(
	ctx context.Context, backend *store.BackendState,
) (*store.BackendState, error) {
	payload, err := json.Marshal(backend)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(payload)
	req, err := http.NewRequestWithContext(
		ctx, "POST", c.apiURL("backends", nil), body)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, APIErrorFromResponse(res)
	}
	backend = &store.BackendState{}
	err = readJSONResponse(res, backend)
	return backend, err
}

// BackendUpdate updates the backend
func (c *JWTClient) BackendUpdate(
	ctx context.Context, backend *store.BackendState,
) (*store.BackendState, error) {
	payload, err := json.Marshal(backend)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(payload)
	req, err := http.NewRequestWithContext(
		ctx, "PATCH", c.apiURL("backends/"+backend.ID, nil), body)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, APIErrorFromResponse(res)
	}
	backend = &store.BackendState{}
	err = readJSONResponse(res, backend)
	return backend, err
}

// BackendDelete removes a backend from the cluster
func (c *JWTClient) BackendDelete(
	ctx context.Context, backend *store.BackendState, query url.Values,
) (*store.BackendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, "DELETE", c.apiURL("backends/"+backend.ID, query), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, APIErrorFromResponse(res)
	}
	backend = &store.BackendState{}
	err = readJSONResponse(res, backend)
	return backend, err
}

// BackendMeetingsList retrieves all meetings for a given backend
func (c *JWTClient) BackendMeetingsList(
	ctx context.Context, backendID string, query url.Values,
) ([]*store.MeetingState, error) {
	if query == nil {
		query = url.Values{}
	}
	query.Set("backend_id", backendID)

	req, err := http.NewRequestWithContext(
		ctx, "GET", c.apiURL("meetings", query), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, APIErrorFromResponse(res)
	}
	meetings := []*store.MeetingState{}
	err = readJSONResponse(res, meetings)
	return meetings, err
}

// BackendMeetingsEnd ends all meetings on a given backend
func (c *JWTClient) BackendMeetingsEnd(
	ctx context.Context, backendID string,
) (*store.Command, error) {
	query := url.Values{}
	query.Set("backend_id", backendID)

	req, err := http.NewRequestWithContext(
		ctx, "DELETE", c.apiURL("meetings", query), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, APIErrorFromResponse(res)
	}
	cmd := &store.Command{}
	err = readJSONResponse(res, cmd)
	return cmd, err
}
