package cluster

import (
	"log"

	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// The Controller interfaces with the state of the cluster
// providing methods for retrieving cluster backends and
// frontends.
//
// The controller subscribes to commands.
type Controller struct {
	conn   *pgxpool.Pool
	state  *store.ClusterState
	client *bbb.Client
}

// NewController will initialize the cluster controller
// with a database connection. A BBB client will be created
// which will be used by the backend instances.
func NewController(state *store.ClusterState, conn *pgxpool.Pool) *Controller {
	return &Controller{
		state:  state,
		conn:   conn,
		client: bbb.NewClient(),
	}
}

// Start the controller
func (c *Controller) Start() {
	log.Println("Starting cluster controller")
}

// GetBackendsWithState retrives backends with
// a specific state
func (c *Controller) GetBackendsWithState(
	state string,
) ([]*Backend, error) {
	return nil, nil
}

// GetBackendByID retrievs a specific backend by ID
func (c *Controller) GetBackendByID(id string) (*Backend, error) {
	return nil, nil
}

// GetBackendByHost retrievs a specific backend
// by the unique host name
func (c *Controller) GetBackendByHost(host string) (*Backend, error) {
	return nil, nil
}

// GetBackendByMeetingID a backend associated with a meeting
func (c *Controller) GetBackendByMeetingID(
	m *bbb.Meeting,
) (*Backend, error) {
	return nil, nil
}

// SetBackendForMeeting associates a meeting with a backend

// Commands
