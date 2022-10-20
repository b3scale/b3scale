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
	Available  bool      `json:"available"`
	Database   string    `json:"database"`
	Migrated   bool      `json:"migrated"`
	Version    int       `json:"version"`
	Error      error     `json:"error"`
	MigratedAt time.Time `json:"migrated_at"`
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
	conn, err := m.Connect(ctx, "template1")
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
	start int,
) error {
	if start > len(m.migrations) {
		return fmt.Errorf("no migrations after version: %d", start)
	}
	log.Info().
		Str("name", db).
		Msg("migrating database")

	conn, err := m.Connect(ctx, db)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	migrations := m.migrations[start:]
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
		return &Status{
			Database: m.DB,
			Error:    err,
		}
	}
	status := &Status{
		Available: true,
		Database:  m.DB,
	}
	state, err := MigrationStateFromDB(ctx, conn)
	fmt.Println(state)
	fmt.Println(err)

	return status
}
