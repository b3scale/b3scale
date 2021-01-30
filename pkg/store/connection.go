package store

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
)

// ConnectOpts database connection options
type ConnectOpts struct {
	URL      string
	MaxConns int32
	MinConns int32
}

// Connect establishes a database connection and
// checks the schema version of the database.
func Connect(opts *ConnectOpts) (*pgxpool.Pool, error) {
	log.Debug().Str("url", opts.URL).Msg("using database")

	// Initialize postgres connection
	cfg, err := pgxpool.ParseConfig(opts.URL)
	if err != nil {
		return nil, err
	}

	cfg.ConnConfig.RuntimeParams["application_name"] = os.Args[0]
	if opts.MaxConns == 0 {
		return nil, fmt.Errorf("MaxConns not configured for connection")
	}

	// We need some more connections
	cfg.MaxConns = opts.MaxConns
	cfg.MinConns = opts.MinConns

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
	var (
		current   int
		appliedAt time.Time
	)

	ctx, cancel := context.WithTimeout(
		context.Background(), time.Second)
	defer cancel()

	qry := `
		SELECT version, applied_at
		  FROM __meta__
		 ORDER BY version DESC
		 LIMIT 1
	`
	err := pool.QueryRow(ctx, qry).Scan(&current, &appliedAt)
	if err != nil {
		return err
	}

	log.Info().
		Int("version", current).
		Time("appliedAt", appliedAt).
		Msg("checking database schema")

	if current != version {
		return fmt.Errorf(
			"unexpected database version: %d, required: %d",
			current, version)
	}
	return nil
}
