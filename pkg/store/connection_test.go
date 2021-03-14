package store

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v4"
)

// connectTest to pgx db pool. Use b3scale defaults if
// environment variable is not set.
func connectTest() error {
	url := os.Getenv("B3SCALE_TEST_DB_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5432/b3scale_test"
	}
	return Connect(&ConnectOpts{
		URL:      url,
		MinConns: 2,
		MaxConns: 16})
}

func beginTest(ctx context.Context, t *testing.T) pgx.Tx {
	if pool == nil {
		if err := connectTest(); err != nil {
			panic(err)
		}
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Error(err)
		return nil
	}
	return tx
}
