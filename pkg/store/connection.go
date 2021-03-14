package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
)

var (
	// ErrNotInitialized will be returned if the
	// database pool is accessed before initializing using
	// Connect.
	ErrNotInitialized = errors.New("store not initialized, pool not ready")

	// ErrMaxConnsUnconfigured will be returned, if the
	// the maximum connections are zero.
	ErrMaxConnsUnconfigured = errors.New("MaxConns not configured")
)

// Pool is the stores global connection pool and
// will be initialized during Connect.
// Database transactions can then be started with store.Begin.
var pool *pgxpool.Pool

// ConnectOpts database connection options
type ConnectOpts struct {
	URL      string
	MaxConns int32
	MinConns int32
}

// Connect initializes the connection pool and
// checks the schema version of the database.
func Connect(opts *ConnectOpts) error {
	log.Debug().Str("url", opts.URL).Msg("using database")

	// Initialize postgres connection
	cfg, err := pgxpool.ParseConfig(opts.URL)
	if err != nil {
		return err
	}

	cfg.ConnConfig.RuntimeParams["application_name"] = os.Args[0]
	if opts.MaxConns == 0 {
		return ErrMaxConnsUnconfigured
	}

	// We need some more connections
	cfg.MaxConns = opts.MaxConns
	cfg.MinConns = opts.MinConns

	p, err := pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		return err
	}
	if err = AssertDatabaseVersion(p, 1); err != nil {
		return err
	}

	// Use pool
	pool = p

	return nil
}

// Acquire tries to get a database connection
// from the pool
func Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	if pool == nil {
		return nil, ErrNotInitialized
	}
	return pool.Acquire(ctx)
}

// begin starts a transaction in the database pool.
func begin(ctx context.Context) (pgx.Tx, error) {
	if pool == nil {
		return nil, ErrNotInitialized
	}
	return pool.Begin(ctx)
}

// beginFunc executes a function with a transaction and
// will forward the error. Rollbacks and commits will
// be handled.
func beginFunc(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
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
