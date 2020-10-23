package cluster

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

// The State of the cluster holds the current backends
// and frontends in the cluster.
type State struct {
	conn *pgxpool.Pool
}

// NewState will initialize the cluster state
// with a database connection.
func NewState(conn *pgxpool.Pool) *State {
	return &State{
		conn: conn,
	}
}

// GetBackendsOpts provides filtering options
// for the GetBackends operation
type GetBackendsOpts struct {
	FilterState string
}

// GetBackends retrives backends
func (c *State) GetBackends(opts *GetBackendsOpts) ([]*Backend, error) {
	return nil, nil
}

// GetBackendByID retrievs a specific backend by ID
func (c *State) GetBackendByID(id string) (*Backend, error) {
	return nil, nil
}

// GetBackendByHost retrievs a specific backend
// by the unique host name
func (c *State) GetBackendByHost(host string) (*Backend, error) {
	return nil, nil
}
