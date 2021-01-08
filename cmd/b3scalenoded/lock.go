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
	ErrBackendNotFound       = errors.New("backend not found in cluster")
	ErrBackendState          = errors.New("backend is not ready")
	ErrBackendDecommissioned = errors.New("backend is decommissioned")
)

// Acquire a lock on the backend to mark the presence
// of the noded. This should be done in a goroutine, as we
// begin a transaction which we will never commit and keep
// open for as long as we are alive.
func acquireBackendNodeLock(pool *pgxpool.Pool, backend *store.BackendState) {
	// We shall never return
	for {
		err := _acquireBackendNodeLock(pool, backend)
		if errors.Is(err, ErrBackendDecommissioned) {
			log.Info().
				Str("backendID", backend.ID).
				Str("host", backend.Backend.Host).
				Msg("backend decommissioned; not longer trying to get a lock")
			return
		}
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
	defer tx.Rollback(ctx)

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

	for {
		// Check if we should loose the lock
		var (
			nodeState  string
			adminState string
		)

		// We are using the pool here to prevent some form
		// of transaction isolation - but maybe I'm too paranoid here.
		err := pool.QueryRow(ctx, `
			SELECT node_state, admin_state
			  FROM backends
			 WHERE id = $1`, backend.ID).Scan(&nodeState, &adminState)
		if err != nil {
			return err
		}

		if nodeState != "ready" && nodeState != "init" {
			log.Warn().
				Str("nodeState", nodeState).
				Str("expected", "init|ready").
				Msg("unexpected node state - releasing lock")
			return ErrBackendState
		}

		// Check if there are meetings alive if not and the
		// desired next state is decommissioned, we can drop the lock
		if adminState == "decommissioned" {
			mstates, err := store.GetMeetingStates(pool, store.Q().
				Where("meetings.backend_id = ?", backend.ID).
				Where("meetings.state->'Running' = ?", true))
			if err != nil {
				return err
			}
			if len(mstates) == 0 {
				return ErrBackendDecommissioned
			}
		}

		// Keep alive
		_, err = tx.Exec(ctx, "SELECT 1")
		if err != nil {
			return err
		}

		time.Sleep(1 * time.Second)
	}
}
