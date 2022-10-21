package store

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4"
)

func beginTest(ctx context.Context, t *testing.T) pgx.Tx {
	if pool == nil {
		panic("pool not initialized")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Error(err)
		return nil
	}
	return tx
}
