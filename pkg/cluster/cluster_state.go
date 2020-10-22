package cluster

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

// The State of the cluster holds the current backends
// and frontends in the cluster.
type State struct {
	conn *pgxpool.Conn
}

// NewState will initialize the cluster state
// with a database connection.
func NewState(conn *pgxpool.Conn) *ClusterState {
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
func (c *ClusterState) GetBackends(opts *GetBackendsOpts) {

}

// GetBackend .... . .. . ..  provides filtering options
// for the GetBackend operation
func (c *ClusterState) GetBackend() {

}
