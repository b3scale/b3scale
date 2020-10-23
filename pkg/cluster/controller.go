package cluster

import (
	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The State of the cluster holds the current backends
// and frontends in the cluster.
type Controller struct {
	conn   *pgxpool.Pool
	client *bbb.Client
}

// NewController will initialize the cluster controller
// with a database connection. A BBB client will be created
// which will be used by the backend instances.
func NewController(conn *pgxpool.Pool) *State {
	return &State{
		conn:   conn,
		client: bbb.NewClient(),
	}
}

// GetBackendsOpts provides filtering options
// for the GetBackends operation
type GetBackendsOpts struct {
	FilterState string
}

// GetBackends retrives backends
func (c *Controller) GetBackends(opts *GetBackendsOpts) ([]*Backend, error) {
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
