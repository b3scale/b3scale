package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Connect establishes a database connection and
// checks the schema version of the database.
func Connect(url string) (*pgxpool.Pool, error) {
	// Initialize postgres connection
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	// We need some more connections
	cfg.MaxConns = 256
	cfg.MinConns = 8

	pool, err := pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	if err = AssertDatabaseVersion(pool, 1); err != nil {
		return nil, err
	}

	return pool, nil
}

// AssertDatabaseVersion tests if the current
// version of the database is equal to a required version
func AssertDatabaseVersion(pool *pgxpool.Pool, version int) error {
	var current int
	ctx := context.Background()
	qry := `SELECT MAX(version) FROM __meta__`
	err := pool.QueryRow(ctx, qry).Scan(&current)
	if err != nil {
		return err
	}

	if current != version {
		return fmt.Errorf(
			"unexpected database version: %d, required: %d",
			current, version)
	}
	return nil
}
