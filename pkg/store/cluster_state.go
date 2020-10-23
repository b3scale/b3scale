package store

import (
	"github.com/jackc/pgx/v4/pgxpool"
	// "gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The ClusterState provides a shared state of the
// cluster.
type ClusterState struct {
	conn *pgxpool.Pool
}

// NewClusterState makes a new cluster state instance
func NewClusterState(conn *pgxpool.Pool) *ClusterState {
	return &ClusterState{
		conn: conn,
	}
}
