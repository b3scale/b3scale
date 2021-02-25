package store

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The FrontendState holds shared information about
// a frontend.
type FrontendState struct {
	ID string

	Active   bool
	Frontend *bbb.Frontend

	CreatedAt time.Time
	UpdatedAt time.Time
}

// InitFrontendState initializes the state with a
// database pool and default values where required.
func InitFrontendState(init *FrontendState) *FrontendState {
	if init.Frontend == nil {
		init.Frontend = &bbb.Frontend{}
	}
	return init
}

// GetFrontendStates retrievs all frontend states from
// the database.
func GetFrontendStates(
	ctx context.Context,
	q sq.SelectBuilder,
) ([]*FrontendState, error) {
	tx := MustTransactionFromContext(ctx)

	qry, params, _ := q.Columns(
		"id",
		"key",
		"secret",
		"active",
		"created_at",
		"updated_at").
		From("frontends").
		ToSql()
	rows, err := tx.Query(ctx, qry, params...)
	if err != nil {
		return nil, err
	}

	// Load and decode results
	cmd := rows.CommandTag()
	results := make([]*FrontendState, 0, cmd.RowsAffected())
	for rows.Next() {
		state := InitFrontendState(&FrontendState{})
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
	ctx context.Context,
	q sq.SelectBuilder,
) (*FrontendState, error) {
	states, err := GetFrontendStates(ctx, q)
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		return nil, nil
	}
	return states[0], nil
}

// Save will create or update a frontend state
func (s *FrontendState) Save(ctx context.Context) error {
	if s.CreatedAt.IsZero() {
		return s.insert(ctx)
	}
	return s.update(ctx)
}

// insert will create a new row with the frontend
// state in the database
func (s *FrontendState) insert(ctx context.Context) error {
	tx := MustTransactionFromContext(ctx)
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
	if err := tx.QueryRow(ctx, qry,
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
func (s *FrontendState) update(ctx context.Context) error {
	tx := MustTransactionFromContext(ctx)
	s.UpdatedAt = time.Now().UTC()
	qry := `
		UPDATE frontends
		   SET key        = $2,
		       secret     = $3,
			   active     = $4,
			   updated_at = $5
		 WHERE id = $1`
	if _, err := tx.Exec(ctx, qry,
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

// Delete will remove a frontend state from the store
func (s *FrontendState) Delete(ctx context.Context) error {
	tx := MustTransactionFromContext(ctx)
	qry := `
		DELETE FROM frontends WHERE id = $1
	`
	_, err := tx.Exec(ctx, qry, s.ID)
	return err
}
