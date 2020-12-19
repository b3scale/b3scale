package main

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// Errors
var (
	ErrBackendNotFound = errors.New("backend not found in cluster")
)

// Acquire a lock on the backend to mark the presence
// of the noded. This should be done in a goroutine, as we
// begin a transaction which we will never commit and keep
// open for as long as we are alive.
func acquireBackendNodeLock(pool *pgxpool.Pool, backend *store.BackendState) {
	// We shall never return
	for {
		err := _acquireBackendNodeLock(pool, backend)
		log.Error().
			Str("backendID", backend.ID).
			Str("host", backend.Backend.Host).
			Err(err).
			Msg("lost lock on backend; trying to reacquire")

		time.Sleep(1 * time.Second)
	}
}

// Internal: acquire the backend lock
func _acquireBackendNodeLock(pool *pgxpool.Pool, backend *store.BackendState) error {
	// Query for locking the backend id in the offline indicator table
	qry := `
		SELECT backend_id
		  FROM backends_node_offline
		 WHERE backend_id = $1
		   FOR UPDATE
	`
	ctx := context.Background() // We do not want any timeouts here
	tx, err := pool.Begin(ctx)  // Begin the transaction
	if err != nil {
		return err
	}

	// Acquire lock
	res, err := tx.Exec(ctx, qry, backend.ID)
	if err != nil {
		return err
	}

	if res.RowsAffected() > 0 {
		log.Info().
			Str("backendID", backend.ID).
			Str("host", backend.Backend.Host).
			Msg("successfully acquired backend node lock")
	} else {
		return ErrBackendNotFound
	}

	// Keep alive
	for {
		_, err := tx.Exec(ctx, "SELECT 1")
		if err != nil {
			return err
		}

		time.Sleep(1 * time.Second)
	}
}
