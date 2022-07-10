package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/b3scale/b3scale/pkg/store"
)

// Client is an interface to the v1 API.
type Client interface {
	Status(ctx context.Context) (*StatusResponse, error)

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
	FrontendUpdateRaw(
		ctx context.Context, id string, payload []byte,
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
	BackendUpdateRaw(
		ctx context.Context, id string, payload []byte,
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

	*http.Client
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

// Build the request URL by joining the API base with the
// api path and resource.
func (c *JWTClient) apiURL(resource string, query url.Values) string {
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
func (c *JWTClient) AuthorizeRequest(req *http.Request) *http.Request {
	bearer := "Bearer " + c.AccessToken
	req.Header.Set("Authorization", bearer)
	return req
}

// Status retrievs the API / server status
func (c *JWTClient) Status(
	ctx context.Context,
) (*StatusResponse, error) {
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
		return nil, APIErrorFromResponse(res)
	}
	status := &StatusResponse{}
	err = readJSONResponse(res, status)
	return status, err
}

// FrontendsList retrievs a list of frontends
func (c *JWTClient) FrontendsList(
	ctx context.Context, query url.Values,
) ([]*store.FrontendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, c.apiURL("frontends", query), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(c.AuthorizeRequest(req))
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, APIErrorFromResponse(res)
	}
	frontends := []*store.FrontendState{}
	err = readJSONResponse(res, &frontends)
	return frontends, err
}

// FrontendRetrieve retrieves a single frontend
func (c *JWTClient) FrontendRetrieve(
	ctx context.Context, id string,
) (*store.FrontendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, c.apiURL("frontends/"+id, nil), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(c.AuthorizeRequest(req))
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
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
	payload, err := json.Marshal(frontend)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(payload)
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, c.apiURL("frontends", nil), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Client.Do(c.AuthorizeRequest(req))
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, APIErrorFromResponse(res)
	}
	frontend = &store.FrontendState{}
	err = readJSONResponse(res, frontend)
	return frontend, err
}

// FrontendUpdateRaw PATCHes an already existing frontend
// identified by ID using raw payload.
func (c *JWTClient) FrontendUpdateRaw(
	ctx context.Context,
	id string,
	payload []byte,
) (*store.FrontendState, error) {
	body := bytes.NewBuffer(payload)
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPatch, c.apiURL("frontends/"+id, nil), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Client.Do(c.AuthorizeRequest(req))
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

// FrontendUpdate PATCHes an already existing frontend.
func (c *JWTClient) FrontendUpdate(
	ctx context.Context, frontend *store.FrontendState,
) (*store.FrontendState, error) {
	payload, err := json.Marshal(frontend)
	if err != nil {
		return nil, err
	}
	return c.FrontendUpdateRaw(ctx, frontend.ID, payload)
}

// FrontendDelete removes a frontend from the cluster.
func (c *JWTClient) FrontendDelete(
	ctx context.Context, frontend *store.FrontendState,
) (*store.FrontendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodDelete, c.apiURL("frontends/"+frontend.ID, nil), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(c.AuthorizeRequest(req))
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, APIErrorFromResponse(res)
	}
	frontend = &store.FrontendState{}
	err = readJSONResponse(res, frontend)
	return frontend, err
}

// BackendsList retrievs a list of backends from the server
func (c *JWTClient) BackendsList(
	ctx context.Context, query url.Values,
) ([]*store.BackendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, c.apiURL("backends", query), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(c.AuthorizeRequest(req))
	if err != nil {
		return nil, err
	}
	if !httpSuccess(res) {
		return nil, APIErrorFromResponse(res)
	}
	backends := []*store.BackendState{}
	err = readJSONResponse(res, &backends)
	return backends, err
}

// BackendRetrieve retrieves a single backend by ID.
func (c *JWTClient) BackendRetrieve(
	ctx context.Context, id string,
) (*store.BackendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodGet, c.apiURL("backends/"+id, nil), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(c.AuthorizeRequest(req))
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
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
		ctx, http.MethodPost, c.apiURL("backends", nil), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Client.Do(c.AuthorizeRequest(req))
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

// BackendUpdateRaw updates an existing backend
// identified by ID with a raw JSON payload.
func (c *JWTClient) BackendUpdateRaw(
	ctx context.Context,
	id string,
	payload []byte,
) (*store.BackendState, error) {
	body := bytes.NewBuffer(payload)
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPatch, c.apiURL("backends/"+id, nil), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := c.Client.Do(c.AuthorizeRequest(req))
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

// BackendUpdate updates the backend
func (c *JWTClient) BackendUpdate(
	ctx context.Context, backend *store.BackendState,
) (*store.BackendState, error) {
	payload, err := json.Marshal(backend)
	if err != nil {
		return nil, err
	}
	return c.BackendUpdateRaw(ctx, backend.ID, payload)
}

// BackendDelete removes a backend from the cluster
func (c *JWTClient) BackendDelete(
	ctx context.Context, backend *store.BackendState, query url.Values,
) (*store.BackendState, error) {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodDelete, c.apiURL("backends/"+backend.ID, query), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(c.AuthorizeRequest(req))
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
		ctx, http.MethodGet, c.apiURL("meetings", query), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(c.AuthorizeRequest(req))
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
		ctx, http.MethodDelete, c.apiURL("meetings", query), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Client.Do(c.AuthorizeRequest(req))
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
