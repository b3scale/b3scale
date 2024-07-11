package store

import (
	"context"
	"errors"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"

	"github.com/b3scale/b3scale/pkg/bbb"
)

// The FrontendState holds shared information about
// a frontend.
type FrontendState struct {
	ID string `json:"id"`

	Active   bool          `json:"active" doc:"When false, the frontend can not longer use the API."`
	Frontend *bbb.Frontend `json:"bbb" api:"FrontendConfig"`

	Settings FrontendSettings `json:"settings"`

	AccountRef *string `json:"account_ref" doc:"If not null, the frontend is bound to an account reference. The reference is freeform string. It is recommended to encode it as base64, but this is optional."`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// InitFrontendState initializes the state with a
// database pool and default values where required.
func InitFrontendState(init *FrontendState) *FrontendState {
	if init.Frontend == nil {
		init.Frontend = &bbb.Frontend{}
	}
	init.Active = true
	return init
}

// GetFrontendStates retrieves all frontend states from
// the database.
func GetFrontendStates(
	ctx context.Context,
	tx pgx.Tx,
	q sq.SelectBuilder,
) ([]*FrontendState, error) {
	qry, params, _ := q.Columns(
		"frontends.id",
		"frontends.key",
		"frontends.secret",
		"frontends.active",
		"frontends.settings",
		"frontends.account_ref",
		"frontends.created_at",
		"frontends.updated_at").
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
			&state.Settings,
			&state.AccountRef,
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
	tx pgx.Tx,
	q sq.SelectBuilder,
) (*FrontendState, error) {
	states, err := GetFrontendStates(ctx, tx, q)
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		return nil, nil
	}
	return states[0], nil
}

// Save will create or update a frontend state
func (s *FrontendState) Save(
	ctx context.Context,
	tx pgx.Tx,
) error {
	if s.CreatedAt.IsZero() {
		return s.insert(ctx, tx)
	}
	return s.update(ctx, tx)
}

// insert will create a new row with the frontend
// state in the database
func (s *FrontendState) insert(ctx context.Context, tx pgx.Tx) error {
	qry := `
		INSERT INTO frontends (
			key, secret, active, settings, account_ref
		) VALUES (
			$1, $2, $3, $4, $5
		)
		RETURNING id, created_at`

	var (
		id        string
		createdAt time.Time
	)
	if err := tx.QueryRow(ctx, qry,
		s.Frontend.Key,
		s.Frontend.Secret,
		s.Active,
		s.Settings,
		s.AccountRef).Scan(&id, &createdAt); err != nil {
		return err
	}
	// Update local state
	s.ID = id
	s.CreatedAt = createdAt
	return nil
}

// update a database row of a frontend state
func (s *FrontendState) update(ctx context.Context, tx pgx.Tx) error {
	s.UpdatedAt = time.Now().UTC()
	qry := `
		UPDATE frontends
		   SET key         = $2,
		       secret      = $3,
			   active      = $4,
			   settings    = $5,
			   account_ref = $6,
			   updated_at  = $7
		 WHERE id = $1`
	if _, err := tx.Exec(ctx, qry,
		s.ID,
		// Values
		s.Frontend.Key,
		s.Frontend.Secret,
		s.Active,
		s.Settings,
		s.AccountRef,
		s.UpdatedAt); err != nil {
		return err
	}
	return nil
}

// Delete will remove a frontend state from the store
func (s *FrontendState) Delete(ctx context.Context, tx pgx.Tx) error {
	qry := `
		DELETE FROM frontends WHERE id = $1
	`
	_, err := tx.Exec(ctx, qry, s.ID)
	return err
}

// Validate checks for presence of required fields.
func (s *FrontendState) Validate() ValidationError {
	err := ValidationError{}

	if s.Frontend == nil {
		err.Add("bbb", "this field is required")
		return err
	}

	s.Frontend.Key = strings.TrimSpace(s.Frontend.Key)
	s.Frontend.Secret = strings.TrimSpace(s.Frontend.Secret)

	if s.Frontend.Key == "" {
		err.Add("bbb.key", ErrFieldRequired)
	}
	if s.Frontend.Secret == "" {
		err.Add("bbb.secret", ErrFieldRequired)
	}

	if len(err) > 0 {
		return err
	}
	return nil
}

// LookupFrontendIDByMeetingID queries the frontend_meetings
// mapping and returns the frontendID for a given meetingID.
// The function name might be a hint.
func LookupFrontendIDByMeetingID(
	ctx context.Context,
	tx pgx.Tx,
	meetingID string,
) (string, bool, error) {
	var frontendID string

	qry := `
		SELECT frontend_id FROM frontend_meetings
		 WHERE meeting_id = $1
	`

	err := tx.QueryRow(ctx, qry, meetingID).Scan(&frontendID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}

	return frontendID, true, nil
}

// RemoveStaleFrontendMeetings removes all frontend
// meetings older than a threshold.
func RemoveStaleFrontendMeetings(
	ctx context.Context,
	tx pgx.Tx,
	t time.Time,
) error {
	qry := `
		DELETE FROM frontend_meetings
		 WHERE seen_at < $1
	`
	_, err := tx.Exec(ctx, qry, t)
	return err
}
