package schema

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"
)

// QryVersionUpdate inserts the current version and description
// into the meta table
const QryVersionUpdate = `
  INSERT INTO __meta__ (version, description) VALUES ($1, $2)
`

// MigrationState is the current version of the database
// and when it was applied.
type MigrationState struct {
	AppliedAt   time.Time `json:"applied_at"`
	Description string    `json:"description"`
	Version     int       `json:"version"`
}

// MigrationStateFromDB reads the current migration state
// from the meta table.
func MigrationStateFromDB(ctx context.Context, conn *pgx.Conn) (*MigrationState, error) {
	sql := `
		SELECT version, description, applied_at
		  FROM __meta__
		ORDER BY version DESC
		LIMIT 1
	`
	state := &MigrationState{}
	if err := conn.QueryRow(ctx, sql).Scan(
		&state.Version,
		&state.Description,
		&state.AppliedAt,
	); err != nil {
		return nil, fmt.Errorf("database not migrated")
	}

	return state, nil
}

// Status is the current status of the database
type Status struct {
	Available         bool            `json:"available"`
	Database          string          `json:"database"`
	Migrated          bool            `json:"migrated"`
	Migration         *MigrationState `json:"migration"`
	PendingMigrations int             `json:"pending_migrations"`
	Error             *string         `json:"error"`
}

// Manager is a migration manager
type Manager struct {
	DB         string
	migrations Migrations
	dbURL      string
}

// NewManager creates a new migration manager
func NewManager(dbURL string) *Manager {
	db := dbURL[(strings.LastIndex(dbURL, "/") + 1):]
	dbURL = dbURL[:(strings.LastIndex(dbURL, "/"))]
	// Strip the database from the URL, and use as base
	return &Manager{
		migrations: GetMigrations(),
		dbURL:      dbURL,
		DB:         db,
	}
}

// Connect acquires a connection to a database by name
func (m *Manager) Connect(
	ctx context.Context,
	db string,
) (*pgx.Conn, error) {
	return pgx.Connect(ctx, m.dbURL+"/"+db)
}

// ClearDatabase drops and creates a database
func (m *Manager) ClearDatabase(ctx context.Context, db string) error {
	conn, err := m.Connect(ctx, "")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, `DROP DATABASE IF EXISTS `+db)
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, `CREATE DATABASE `+db)
	if err != nil {
		return err
	}
	return nil
}

// Migrate applies all migrations above the starting version
func (m *Manager) Migrate(
	ctx context.Context,
	db string,
) error {
	conn, err := m.Connect(ctx, db)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	// Get current version
	version := 0
	state, err := MigrationStateFromDB(ctx, conn)
	if err == nil {
		version = state.Version
	}
	if len(m.migrations) <= version {
		log.Info().Msg("database already migrated")
		return nil
	}

	log.Info().
		Str("name", db).
		Int("from_version", version).
		Msg("migrating database")

	migrations := m.migrations[version:]

	for _, mig := range migrations {
		log.Info().
			Int("version", mig.Seq).
			Str("name", mig.Name).
			Str("database", db).
			Msg("applying migration")

		// Apply schema and update metadata
		if _, err := conn.Exec(ctx, mig.SQL); err != nil {
			return err
		}
		if _, err := conn.Exec(ctx, QryVersionUpdate, mig.Seq, mig.Name); err != nil {
			return err
		}
	}

	return nil
}

// Status retrievs the information about the state of the db
func (m *Manager) Status(
	ctx context.Context,
) *Status {
	// Try to connect to primary database
	conn, err := m.Connect(ctx, m.DB)
	if err != nil {
		errStr := err.Error()
		return &Status{
			Database: m.DB,
			Error:    &errStr,
		}
	}
	status := &Status{
		Available: true,
		Database:  m.DB,
	}
	state, err := MigrationStateFromDB(ctx, conn)
	if err != nil {
		log.Error().Err(err).Msg("could not get migration state")
		status.PendingMigrations = len(m.migrations)
	} else {
		status.Migration = state
		status.PendingMigrations = len(m.migrations) - state.Version
		status.Migrated = status.PendingMigrations == 0
	}

	return status
}
