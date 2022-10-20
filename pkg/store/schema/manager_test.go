package schema

import (
	"context"
	"testing"

	"github.com/b3scale/b3scale/pkg/config"
)

func TestNewManager(t *testing.T) {
	m := NewManager(config.EnvDbURLDefault)
	if m.DB != "b3scale" {
		t.Error("unexpected dbName:", m.DB)
	}
	t.Log("URL:", m.dbURL, "Name:", m.DB)
}

func TestMigrationStateFromDB(t *testing.T) {
	ctx := context.Background()
	m := NewManager(config.EnvDbURLDefault)
	db := m.DB + "_testing"
	if err := m.ClearDatabase(ctx, db); err != nil {
		t.Fatal(err)
	}
	conn, err := m.Connect(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	_, err = MigrationStateFromDB(ctx, conn)
	if err == nil {
		t.Fatal("Err should: database not migrated")
	}
	conn.Close(ctx)

	// Apply Migrations
	if err := m.Migrate(ctx, db, 0); err != nil {
		t.Fatal(err)
	}

	conn, err = m.Connect(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	state, err := MigrationStateFromDB(ctx, conn)
	if err != nil {
		t.Fatal("Err should: database not migrated")
	}

	if state.Version == 0 {
		t.Error("expected database version > 0")
	}

	t.Log("migration:", state.Description, "version:", state.Version)
}
