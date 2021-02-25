package store

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	pool *pgxpool.Pool
)

func init() {
	var err error
	pool, err = connectTest()
	if err != nil {
		panic(err)
	}
}

// connectTest to pgx db pool. Use b3scale defaults if
// environment variable is not set.
func connectTest() (*pgxpool.Pool, error) {
	url := os.Getenv("B3SCALE_TEST_DB_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5432/b3scale_test"
	}
	return Connect(&ConnectOpts{
		URL:      url,
		MinConns: 2,
		MaxConns: 16})
}

type endTestFunc func() error

func beginTest(t *testing.T) (context.Context, endTestFunc) {
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Error(err)
	}
	ctx = ContextWithTransaction(ctx, tx)
	end := func() error {
		return tx.Rollback(ctx)
	}
	return ctx, end
}
