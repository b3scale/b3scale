package store

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Connect establishes a database connection and
// checks the schema version of the database.
func Connect(url string) *pgxpool.Pool {
	// Initialize postgres connection
	pool, err := pgxpool.Connect(context.Background(), url)
	if err != nil {
		log.Fatal("Error while connecting to database:", err)
	}
	if err = AssertDatabaseVersion(pool, 1); err != nil {
		log.Fatal(err)
	}

	return pool
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
