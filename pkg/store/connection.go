package store

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Connect establishes a database connection
// and will fataly fail if this is not possible.
func Connect(url string) *pgxpool.Pool {
	// Initialize postgres connection
	conn, err := pgxpool.Connect(context.Background(), url)
	if err != nil {
		log.Fatal("Error while connecting to database:", err)
	}

	return conn
}

// AssertDatabaseVersion tests if the current
// version of the database is equal to a required version
func AssertDatabaseVersion(conn *pgxpool.Pool, version int) error {
	var current int
	ctx := context.Background()
	qry := `SELECT MAX(version) FROM __meta__`
	err := conn.QueryRow(ctx, qry).Scan(&current)
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
