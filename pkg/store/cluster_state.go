package store

import (
	"github.com/jackc/pgx/v4/pgxpool"
)

// The ClusterState provides a shared state of the
// cluster.
type ClusterState struct {
	conn *pgxpool.Pool
}
