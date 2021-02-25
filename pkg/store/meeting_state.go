package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// MeetingStateErrors
var (
	ErrNoBackend = errors.New("no backend associated with meeting")
)

// The MeetingState holds a meeting and it's relations
// to a backend and frontend.
type MeetingState struct {
	ID         string
	InternalID string

	Meeting *bbb.Meeting

	FrontendID *string
	frontend   *FrontendState

	BackendID *string
	backend   *BackendState

	CreatedAt time.Time
	UpdatedAt time.Time
	SyncedAt  time.Time
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
	q sq.SelectBuilder,
) ([]*MeetingState, error) {
	tx := MustTransactionFromContext(ctx)

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
func GetMeetingState(ctx context.Context, q sq.SelectBuilder) (*MeetingState, error) {
	states, err := GetMeetingStates(ctx, q)
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		return nil, nil
	}
	return states[0], nil
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
func (s *MeetingState) GetBackendState(ctx context.Context) (*BackendState, error) {
	if s.BackendID == nil {
		return nil, nil
	}

	if s.backend != nil {
		return s.backend, nil
	}

	// Get related backend state
	var err error
	s.backend, err = GetBackendState(ctx, Q().
		Where("id = ?", s.BackendID))
	if err != nil {
		return nil, err
	}

	return s.backend, nil
}

// GetFrontendState loads the frontend state for the meeting
func (s *MeetingState) GetFrontendState(ctx context.Context) (*FrontendState, error) {
	if s.FrontendID == nil {
		return nil, nil
	}

	if s.frontend != nil {
		return s.frontend, nil
	}

	// Load frontend state from database
	var err error
	s.frontend, err = GetFrontendState(ctx, Q().
		Where("id = ?", s.FrontendID))
	if err != nil {
		return nil, err
	}

	return s.frontend, nil
}

// DeleteMeetingStateByID will remove a meeting state.
// It will succeed, even if no such meeting was present.
// TODO: merge with DeleteMeetingStateByInternalID
func DeleteMeetingStateByID(ctx context.Context, id string) error {
	tx := MustTransactionFromContext(ctx)
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
	if err := updateBackendStatCounters(ctx, *backendID); err != nil {
		return err
	}

	return nil
}

// DeleteMeetingStateByInternalID will remove a meeting state.
// It will succeed, even if no such meeting was present.
func DeleteMeetingStateByInternalID(ctx context.Context, id string) error {
	tx := MustTransactionFromContext(ctx)
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
	if err := updateBackendStatCounters(ctx, *backendID); err != nil {
		return err
	}

	return nil
}

// DeleteOrphanMeetings will remove all meetings not
// in a list of (internal) meeting IDs, but associated
// with a backend
func DeleteOrphanMeetings(
	ctx context.Context,
	backendID string,
	backendMeetings []string,
) (int64, error) {
	tx := MustTransactionFromContext(ctx)

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
func (s *MeetingState) Refresh(ctx context.Context) error {
	// Load from database
	next, err := GetMeetingState(ctx, Q().
		Where("id = ?", s.ID))
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
func (s *MeetingState) Save(ctx context.Context) error {
	var (
		err error
		id  string
	)
	if s.CreatedAt.IsZero() {
		id, err = s.insert(ctx)
		s.ID = id
	} else {
		err = s.update(ctx)
	}
	if err != nil {
		return err
	}

	// Refresh stats if backend is present
	if s.BackendID != nil {
		if err := updateBackendStatCounters(ctx, *s.BackendID); err != nil {
			return err
		}
	}

	return s.Refresh(ctx)
}

// Add a new meeting to the database
func (s *MeetingState) insert(ctx context.Context) (string, error) {
	tx := MustTransactionFromContext(ctx)
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
func (s *MeetingState) update(ctx context.Context) error {
	tx := MustTransactionFromContext(ctx)
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

// UpdateMeetingStateIfExists updates the meeting attribute of a meeting
// The meeting state need to have an internal meeting id.
func UpdateMeetingStateIfExists(ctx context.Context, m *bbb.Meeting) (int64, error) {
	tx := MustTransactionFromContext(ctx)
	internalID := m.InternalMeetingID
	if internalID == "" {
		return 0, fmt.Errorf(
			"can not use meeting for update without InternalMeetingID")
	}

	updatedAt := time.Now().UTC()
	qry := `
		UPDATE meetings
		   SET state		= $2,
			   updated_at   = $3
	 	 WHERE internal_id = $1`
	cmd, err := tx.Exec(ctx, qry, internalID, m, updatedAt)
	if err != nil {
		return 0, err
	}

	return cmd.RowsAffected(), nil
}

// SetBackendID associates a meeting with a backend
func (s *MeetingState) SetBackendID(ctx context.Context, id string) error {
	if s.BackendID != nil && *s.BackendID == id {
		return nil // nothing to do here
	}

	// Bind backend
	s.BackendID = &id
	qry := `UPDATE meetings SET backend_id = $2
			WHERE id = $1`
	tx := MustTransactionFromContext(ctx)
	_, err := tx.Exec(ctx, qry, s.ID, id)
	return err
}

// BindFrontendID associates an unclaimed meeting with a frontend
func (s *MeetingState) BindFrontendID(ctx context.Context, id string) error {
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
	tx := MustTransactionFromContext(ctx)
	qry := `UPDATE meetings SET frontend_id = $2
			WHERE id = $1`
	_, err := tx.Exec(ctx, qry, s.ID, id)
	return err
}

// IsStale checks if the last sync is longer
// ago than a given threashold.
func (s *MeetingState) IsStale() bool {
	return time.Now().UTC().Sub(s.SyncedAt) > 1*time.Minute
}

// MarkSynced sets the synced at timestamp
func (s *MeetingState) MarkSynced() {
	s.SyncedAt = time.Now().UTC()
}
