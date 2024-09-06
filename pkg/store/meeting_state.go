package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/b3scale/b3scale/pkg/bbb"
)

// MeetingStateErrors
var (
	ErrNoBackend        = errors.New("no backend associated with meeting")
	ErrDeadlineRequired = errors.New("the operation requires a dealine")
)

// The MeetingState holds a meeting and its relations
// to a backend and frontend.
type MeetingState struct {
	ID         string `json:"id"`
	InternalID string `json:"internal_id"`

	Meeting *bbb.Meeting `json:"meeting" api:"MeetingInfo"`

	FrontendID *string `json:"frontend_id"`
	frontend   *FrontendState

	BackendID *string `json:"backend_id"`
	backend   *BackendState

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	SyncedAt  time.Time `json:"synced_at"`
}

// InitMeetingState initializes meeting state with
// defaults and a connection
func InitMeetingState(
	init *MeetingState,
) *MeetingState {
	return init
}

// GetMeetingStates retrieves all meeting states
func GetMeetingStates(
	ctx context.Context,
	tx pgx.Tx,
	q sq.SelectBuilder,
) ([]*MeetingState, error) {
	qry, params, _ := q.Columns(
		"meetings.id",
		"meetings.internal_id",
		"meetings.frontend_id",
		"meetings.backend_id",
		"meetings.state",
		"meetings.created_at",
		"meetings.updated_at",
		"meetings.synced_at").
		From("meetings").
		ToSql()
	rows, err := tx.Query(ctx, qry, params...)
	if err != nil {
		return nil, err
	}
	cmd := rows.CommandTag()
	results := make([]*MeetingState, 0, cmd.RowsAffected())
	for rows.Next() {
		state, err := meetingStateFromRow(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, state)
	}
	return results, nil
}

// GetMeetingState tries to retriev a single meeting state
func GetMeetingState(
	ctx context.Context,
	tx pgx.Tx,
	q sq.SelectBuilder,
) (*MeetingState, error) {
	states, err := GetMeetingStates(ctx, tx, q)
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		return nil, nil
	}
	return states[0], nil
}

// AwaitMeetingState polls the database for a meeting
// state until the context expires.
func AwaitMeetingState(
	ctx context.Context,
	conn *pgxpool.Conn,
	q sq.SelectBuilder,
) (*MeetingState, pgx.Tx, error) {
	if _, ok := ctx.Deadline(); !ok {
		return nil, nil, ErrDeadlineRequired
	}

	for {
		if err := ctx.Err(); err != nil {
			return nil, nil, err // context was canceled or expired
		}

		tx, err := conn.Begin(ctx)
		if err != nil {
			time.Sleep(150 * time.Millisecond)
			continue // Let's not give up that easily
		}

		state, err := GetMeetingState(ctx, tx, q)
		if err != nil {
			tx.Rollback(ctx)
			return nil, nil, err // Database error
		}

		if state != nil {
			return state, tx, nil // We are done here!
		}

		// Close tx and wait before retry
		tx.Rollback(ctx)
		time.Sleep(150 * time.Millisecond)
	}
}

// GetMeetingStateByID is a convenience wrapper
// around GetMeetingState.
func GetMeetingStateByID(
	ctx context.Context,
	tx pgx.Tx,
	id string,
) (*MeetingState, error) {
	return GetMeetingState(ctx, tx, Q().
		Where("id = ?", id))
}

func meetingStateFromRow(
	row pgx.Row,
) (*MeetingState, error) {
	state := InitMeetingState(&MeetingState{})
	err := row.Scan(
		&state.ID,
		&state.InternalID,
		&state.FrontendID,
		&state.BackendID,
		&state.Meeting,
		&state.CreatedAt,
		&state.UpdatedAt,
		&state.SyncedAt)
	if err != nil {
		return nil, err
	}

	return state, err
}

// GetBackendState loads the backend state
func (s *MeetingState) GetBackendState(
	ctx context.Context,
	tx pgx.Tx,
) (*BackendState, error) {
	if s.BackendID == nil {
		return nil, nil
	}

	if s.backend != nil {
		return s.backend, nil
	}

	// Get related backend state
	var err error
	s.backend, err = GetBackendState(ctx, tx, Q().
		Where("id = ?", s.BackendID))
	if err != nil {
		return nil, err
	}

	return s.backend, nil
}

// GetFrontendState loads the frontend state for the meeting
func (s *MeetingState) GetFrontendState(
	ctx context.Context,
	tx pgx.Tx,
) (*FrontendState, error) {
	if s.FrontendID == nil {
		return nil, nil
	}

	if s.frontend != nil {
		return s.frontend, nil
	}

	// Load frontend state from database
	var err error
	s.frontend, err = GetFrontendState(ctx, tx, Q().
		Where("id = ?", s.FrontendID))
	if err != nil {
		return nil, err
	}

	return s.frontend, nil
}

// DeleteMeetingStateByID will remove a meeting state.
// It will succeed, even if no such meeting was present.
// TODO: merge with DeleteMeetingStateByInternalID
func DeleteMeetingStateByID(
	ctx context.Context,
	tx pgx.Tx,
	id string,
) error {
	// Get affected backend
	var backendID *string
	qry := `
		SELECT backend_id FROM meetings WHERE id = $1
	`
	if err := tx.
		QueryRow(ctx, qry, id).
		Scan(&backendID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	qry = `
		DELETE FROM meetings WHERE id = $1
	`
	if _, err := tx.Exec(ctx, qry, id); err != nil {
		return err
	}

	if backendID == nil {
		return nil // Meeting is not associated with a backend
	}

	// Update stat counters
	if err := updateBackendStatCounters(ctx, tx, *backendID); err != nil {
		return err
	}

	return nil
}

// DeleteMeetingStateByInternalID will remove a meeting state.
// It will succeed, even if no such meeting was present.
func DeleteMeetingStateByInternalID(
	ctx context.Context,
	tx pgx.Tx,
	id string,
) error {
	// Get affected backend
	var backendID *string
	qry := `
		SELECT backend_id FROM meetings WHERE id = $1
	`
	if err := tx.
		QueryRow(ctx, qry, id).
		Scan(&backendID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	qry = `
		DELETE FROM meetings WHERE internal_id = $1
	`
	if _, err := tx.Exec(ctx, qry, id); err != nil {
		return err
	}

	if backendID == nil {
		return nil // Meeting is not associated with a backend
	}

	// Update stat counters
	if err := updateBackendStatCounters(ctx, tx, *backendID); err != nil {
		return err
	}

	return nil
}

// DeleteOrphanMeetings will remove all meetings not
// in a list of (internal) meeting IDs, but associated
// with a backend
func DeleteOrphanMeetings(
	ctx context.Context,
	tx pgx.Tx,
	backendID string,
	backendMeetings []string,
) (int64, error) {
	// Okay. I tried to do this the nice way... however
	// trying to construct something with 'NOT IN' failed,
	// and the only workaround I found was constructing
	// programmatically a long list of 'AND NOT internal_id = ...'
	//
	// I guess this can be optimized.
	q := NewDelete().
		From("meetings").
		Where("backend_id = ?", backendID)

	for _, id := range backendMeetings {
		q = q.Where("internal_id <> ?", id)
	}
	qry, params, err := q.ToSql()
	if err != nil {
		return 0, err
	}

	cmd, err := tx.Exec(ctx, qry, params...)
	if err != nil {
		return 0, err
	}
	return cmd.RowsAffected(), nil
}

// Refresh the backend state from the database
func (s *MeetingState) Refresh(ctx context.Context, tx pgx.Tx) error {
	next, err := GetMeetingState(ctx, tx, Q().Where("id = ?", s.ID))
	if err != nil {
		return err
	}
	if next == nil {
		return fmt.Errorf("meeting %s is gone", s.ID)
	}
	*s = *next
	return nil
}

// Save updates or inserts a meeting state into our
// cluster state.
func (s *MeetingState) Save(ctx context.Context, tx pgx.Tx) error {
	var (
		err error
		id  string
	)
	if s.CreatedAt.IsZero() {
		id, err = s.insert(ctx, tx)
		s.ID = id
	} else {
		err = s.update(ctx, tx)
	}
	if err != nil {
		return err
	}

	// Refresh stats if backend is present
	if s.BackendID != nil {
		if err := updateBackendStatCounters(ctx, tx, *s.BackendID); err != nil {
			return err
		}
	}

	// Refresh frontend binding
	if err := s.UpdateFrontendMeetingMapping(ctx, tx); err != nil {
		return err
	}

	return s.Refresh(ctx, tx)
}

// Add a new meeting to the database
func (s *MeetingState) insert(ctx context.Context, tx pgx.Tx) (string, error) {
	qry := `
		INSERT INTO meetings (
			id,
			internal_id,

			state,

			frontend_id,
			backend_id
		) VALUES (
			$1, $2, $3, $4, $5
		) RETURNING id`
	err := tx.QueryRow(ctx, qry,
		s.Meeting.MeetingID,
		s.Meeting.InternalMeetingID,
		s.Meeting,
		s.FrontendID,
		s.BackendID).Scan(&s.ID)
	if err != nil {
		return "", err
	}

	return s.ID, nil
}

// Update the meeting state
func (s *MeetingState) update(ctx context.Context, tx pgx.Tx) error {
	s.UpdatedAt = time.Now().UTC()
	qry := `
		UPDATE meetings
		   SET state		= $2,
		       internal_id  = $3,
		       frontend_id  = $4,
			   backend_id   = $5,
		  	   synced_at    = $6,
			   updated_at   = $7 
	 	 WHERE id = $1`
	_, err := tx.Exec(ctx, qry,
		s.ID,
		s.Meeting,
		s.Meeting.InternalMeetingID,
		s.FrontendID,
		s.BackendID,
		s.SyncedAt,
		s.UpdatedAt)
	return err
}

// UpdateFrontendMeetingMapping updates the `frontend_meetings`
// mapping table. This table does not hold any state, but persists the
// association between frontend and meetingID for identifiying recordings.
func (s *MeetingState) UpdateFrontendMeetingMapping(
	ctx context.Context,
	tx pgx.Tx,
) error {
	if s.FrontendID == nil {
		return nil // nothing to do here
	}
	qry := `
		INSERT INTO frontend_meetings (
			frontend_id,
			meeting_id
		) VALUES (
			$1, $2
		) ON CONFLICT (meeting_id) DO UPDATE
		  SET seen_at = CURRENT_TIMESTAMP
	`
	_, err := tx.Exec(ctx, qry, *s.FrontendID, s.ID)
	return err
}

// Upsert meeting state will create the meeting state
// or will fall back to a state update.
func (s *MeetingState) Upsert(ctx context.Context, tx pgx.Tx) (string, error) {
	qry := `
		INSERT INTO meetings (
			id,
			internal_id,

			state,

			frontend_id,
			backend_id,

			updated_at,
			synced_at

		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		ON CONFLICT ON CONSTRAINT meetings_pkey DO UPDATE
		   SET state		= EXCLUDED.state,
		  	   synced_at    = EXCLUDED.synced_at,
			   updated_at   = EXCLUDED.updated_at
		RETURNING id`

	err := tx.QueryRow(ctx, qry,
		s.Meeting.MeetingID,
		s.Meeting.InternalMeetingID,
		s.Meeting,
		s.FrontendID,
		s.BackendID,
		s.UpdatedAt,
		s.SyncedAt).Scan(&s.ID)
	if err != nil {
		return "", err
	}

	return s.ID, nil
}

// SetBackendID associates a meeting with a backend
func (s *MeetingState) SetBackendID(
	ctx context.Context,
	tx pgx.Tx,
	id string,
) error {
	if s.BackendID != nil && *s.BackendID == id {
		return nil // nothing to do here
	}
	// Bind backend
	s.BackendID = &id
	qry := `UPDATE meetings SET backend_id = $2
			WHERE id = $1`
	_, err := tx.Exec(ctx, qry, s.ID, id)
	return err
}

// BindFrontendID associates an unclaimed meeting with a frontend
func (s *MeetingState) BindFrontendID(
	ctx context.Context,
	tx pgx.Tx,
	id string,
) error {
	if s.FrontendID != nil {
		if *s.FrontendID == id {
			// Nothing to do here
			return nil
		}
		// We do not support rebinding right now...
		return fmt.Errorf(
			"meeting is already associated with different frontend")
	}
	// Bind frontend
	s.FrontendID = &id
	qry := `UPDATE meetings SET frontend_id = $2
			WHERE id = $1`
	_, err := tx.Exec(ctx, qry, s.ID, id)
	return err
}

// IsStale checks if the last sync is longer
// ago than a given threshold.
func (s *MeetingState) IsStale(threshold time.Duration) bool {
	return time.Now().UTC().Sub(s.SyncedAt) > threshold
}

// MarkSynced sets the synced at timestamp
func (s *MeetingState) MarkSynced() {
	s.SyncedAt = time.Now().UTC()
}
