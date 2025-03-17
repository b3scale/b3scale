package api

import (
	"context"
	"net/url"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/store"
	"github.com/b3scale/b3scale/pkg/store/schema"
)

// FrontendResourceClient defines methods for using
// the frontends resource of the api
type FrontendResourceClient interface {
	FrontendsList(
		ctx context.Context, query ...url.Values,
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
}

// BackendResourceClient defines methods for
// using the backends api resource
type BackendResourceClient interface {
	BackendsList(
		ctx context.Context, query ...url.Values,
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
		ctx context.Context,
		backend *store.BackendState,
		opts ...url.Values,
	) (*store.BackendState, error)
}

// MeetingResourceClient defines methods for accessing
// the meetings api resource
type MeetingResourceClient interface {
	BackendMeetingsList(
		ctx context.Context,
		backendID string,
		query ...url.Values,
	) ([]*store.MeetingState, error)

	MeetingsList(
		ctx context.Context,
		query ...url.Values,
	) ([]*store.MeetingState, error)
	MeetingRetrieve(
		ctx context.Context,
		id string,
	) (*store.MeetingState, error)
	MeetingUpdateRaw(
		ctx context.Context,
		id string,
		payload []byte,
	) (*store.MeetingState, error)
	MeetingUpdate(
		ctx context.Context,
		meeting *store.MeetingState,
	) (*store.MeetingState, error)
	MeetingDelete(
		ctx context.Context,
		id string,
	) (*store.MeetingState, error)
}

// RecordingsResourceClient defines recording related methods.
type RecordingsResourceClient interface {
	RecordingsListByFrontendID(
		ctx context.Context,
		feID string,
	) ([]*store.RecordingState, error)
	RecordingsListByFrontendKey(
		ctx context.Context,
		feKey string,
	) ([]*store.RecordingState, error)
	RecordingsRetrieve(
		ctx context.Context,
		id string,
	) (*store.RecordingState, error)
	RecordingsSetVisibility(
		ctx context.Context,
		id string,
		v bbb.RecordingVisibility,
	) (*store.RecordingState, error)
}

// CommandResourceClient defines methods for creating
// and polling commands
type CommandResourceClient interface {
	BackendMeetingsEnd(
		ctx context.Context,
		backendID string,
	) (*store.Command, error)

	CommandCreate(
		ctx context.Context,
		cmd *store.Command,
	) (*store.Command, error)
	CommandRetrieve(
		ctx context.Context,
		id string,
	) (*store.Command, error)

	// Control commands
	CtrlMigrate(
		ctx context.Context,
	) (*schema.Status, error)
}

// AgentResourceClient defines node agent specific
// methods.
type AgentResourceClient interface {
	AgentHeartbeatCreate(
		ctx context.Context,
	) (*store.AgentHeartbeat, error)
	AgentBackendRetrieve(
		ctx context.Context,
	) (*store.BackendState, error)
	AgentRPC(
		ctx context.Context,
		req *RPCRequest,
	) (RPCResult, error)
}

// Client is an interface to the api API.
type Client interface {
	Status(ctx context.Context) (*StatusResponse, error)

	FrontendResourceClient
	BackendResourceClient
	MeetingResourceClient
	RecordingsResourceClient
	CommandResourceClient
	AgentResourceClient
}
