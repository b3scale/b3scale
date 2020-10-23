package store

import (
	"os"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
)

// connectTest to pgx db pool. Use b3scale defaults if
// environment variable is not set.
func connectTest(t *testing.T) *pgxpool.Pool {
	url := os.Getenv("B3SCALE_TEST_DB_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5432/b3scale_test"
	}

	conn := Connect(url)

	// Assert current version
	err := AssertDatabaseVersion(conn, 1)
	if err != nil {
		t.Error(err)
		return nil
	}

	return conn
}

func TestConnect(t *testing.T) {
	conn := connectTest(t)
	conn.Close()
}
