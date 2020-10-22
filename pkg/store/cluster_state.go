package store

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

// The ClusterState holds the current backends
// and frontends in the cluster.
type ClusterState struct {
	conn *pgxpool.Conn
}

// NewClusterState will initialize the cluster state
// with a database connection.
func NewClusterState(conn *pgxpool.Conn) *ClusterState {
	return &ClusterState{
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
