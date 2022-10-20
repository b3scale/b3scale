package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/b3scale/b3scale/pkg/store/schema"
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

	cfg.ConnConfig.RuntimeParams["application_name"] = filepath.Base(os.Args[0])
	if opts.MaxConns == 0 {
		return ErrMaxConnsUnconfigured
	}

	// We need some more connections
	cfg.MaxConns = opts.MaxConns
	cfg.MinConns = opts.MinConns

	p, err := pgxpool.ConnectConfig(context.Background(), cfg)
	if err != nil {
		// This error can not be ignored. It is not possible to acquire
		// a connection later, even if the database becomes available.
		return err
	}

	// Use pool
	pool = p
	return nil
}

// ConnectTest to pgx db pool. Use b3scale defaults if
// environment variable is not set.
func ConnectTest(ctx context.Context) error {
	url := os.Getenv("B3SCALE_TEST_DB_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5432/b3scale_test"
	}
	m := schema.NewManager(url)

	// Clear database and apply migrations
	if err := m.ClearDatabase(ctx, m.DB); err != nil {
		return err
	}
	if err := m.Migrate(ctx, m.DB, 0); err != nil {
		return err
	}

	return Connect(&ConnectOpts{
		URL:      url,
		MinConns: 2,
		MaxConns: 16})
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
