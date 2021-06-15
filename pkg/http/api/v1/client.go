package v1

import (
	"context"
	"net/http"
	"net/url"

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

// JWTClient is a http api v1 client
type JWTClient struct {
	Host        string
	AccessToken string
}

// NewJWTClient initializes the client
func NewJWTClient(host, token string) *Client {
	return &JWTClient{
		Host:        host,
		AccessToken: token,
	}
}

// FrontendsList retrievs a list of frontends
func (c *JWTClient) FrontendsList(
	ctx context.Context, query url.Values,
) ([]*store.FrontendState, error) {

	return nil, nil
}

// FrontendRetrieve retrieves a single frontend
func FrontendRetrieve(
	ctx context.Context, id string,
) (*store.FrontendState, error) {
	return nil, nil
}

// FrontendCreate POSTs a new frontend to the server
func FrontendCreate(
	ctx context.Context, frontend *store.FrontendState,
) (*store.FrontendState, error) {
	return nil, nil
}

// FrontendUpdate PATCHes an already existing frontend.
func FrontendUpdate(
	ctx context.Context, frontend *store.FrontendState,
) (*store.FrontendState, error) {
	return nil, nil
}

// FrontendDelete removes a frontend from the cluster.
func FrontendDelete(
	ctx context.Context, frontend *store.FrontendState,
) (*store.FrontendState, error) {
	return nil, nil
}

// BackendsList retrievs a list of backends from the server
func BackendsList(
	ctx context.Context, query url.Values,
) ([]*store.BackendState, error) {
	return nil, nil
}

// BackendRetrieve retrieves a single backend by ID.
func BackendRetrieve(
	ctx context.Context, id string,
) (*store.BackendState, error) {
	return nil, nil
}

// BackendCreate creates a new backend on the server
func BackendCreate(
	ctx context.Context, backend *store.BackendState,
) (*store.BackendState, error) {
	return nil, nil
}

// BackendUpdate updates the backend
func BackendUpdate(
	ctx context.Context, backend *store.BackendState,
) (*store.BackendState, error) {
	return nil, nil
}

// BackendDelete removes a backend from the cluster
func BackendDelete(
	ctx context.Context, backend *store.BackendState,
) (*store.BackendState, error) {
	return nil, nil
}
