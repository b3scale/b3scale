package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The FrontendState holds shared information about
// a frontend.
type FrontendState struct {
	ID string

	Active   bool
	Frontend *bbb.Frontend

	CreatedAt time.Time
	UpdatedAt *time.Time

	pool *pgxpool.Pool
}

// InitFrontendState initializes the state with a
// database pool and default values where required.
func InitFrontendState(pool *pgxpool.Pool, init *FrontendState) *FrontendState {
	init.pool = pool
	if init.Frontend == nil {
		init.Frontend = &bbb.Frontend{}
	}
	return init
}

// GetFrontendStates retrievs all frontend states from
// the database.
func GetFrontendStates(pool *pgxpool.Pool, q *Query) ([]*FrontendState, error) {
	ctx := context.Background()
	qry := `
		SELECT
		  id,

		  key,
		  secret,

		  active,

		  created_at,
		  updated_at
		FROM frontends ` + q.related() + `
		WHERE ` + q.where()
	rows, err := pool.Query(ctx, qry, q.params()...)
	if err != nil {
		return nil, err
	}

	// Load and decode results
	cmd := rows.CommandTag()
	results := make([]*FrontendState, 0, cmd.RowsAffected())
	for rows.Next() {
		state := InitFrontendState(pool, &FrontendState{})
		err := rows.Scan(
			&state.ID,
			&state.Frontend.Key, &state.Frontend.Secret,
			&state.Active,
			&state.CreatedAt, &state.UpdatedAt)
		if err != nil {
			return nil, err
		}
		results = append(results, state)
	}
	return results, nil
}

// GetFrontendState gets a single row from the store.
// This may return nil without an error.
func GetFrontendState(
	pool *pgxpool.Pool,
	q *Query,
) (*FrontendState, error) {
	states, err := GetFrontendStates(pool, q)
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		return nil, nil
	}
	return states[0], nil
}

// Save will create or update a frontend state
func (s *FrontendState) Save() error {
	if s.CreatedAt.IsZero() {
		return s.insert()
	}
	return s.update()
}

// insert will create a new row with the frontend
// state in the database
func (s *FrontendState) insert() error {
	ctx := context.Background()
	qry := `
		INSERT INTO frontends (
			key, secret, active
		) VALUES (
			$1, $2, $3
		)
		RETURNING id, created_at`

	var (
		id        string
		createdAt time.Time
	)
	if err := s.pool.QueryRow(ctx, qry,
		s.Frontend.Key,
		s.Frontend.Secret,
		s.Active).Scan(&id, &createdAt); err != nil {
		return err
	}
	// Update local state
	s.ID = id
	s.CreatedAt = createdAt
	return nil
}

// update a database row of a frontend state
func (s *FrontendState) update() error {
	now := time.Now().UTC()
	s.UpdatedAt = &now
	ctx := context.Background()
	qry := `
		UPDATE frontends
		   SET key        = $2,
		       secret     = $3,
			   active     = $4,
			   updated_at = $5
		 WHERE id = $1`
	if _, err := s.pool.Exec(ctx, qry,
		s.ID,
		// Values
		s.Frontend.Key,
		s.Frontend.Secret,
		s.Active,
		s.UpdatedAt); err != nil {
		return err
	}
	return nil
}
