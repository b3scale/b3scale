package store

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
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

	pool *pgxpool.Pool
}

// InitMeetingState initializes meeting state with
// defaults and a connection
func InitMeetingState(
	pool *pgxpool.Pool,
	init *MeetingState,
) *MeetingState {
	init.pool = pool
	return init
}

// GetMeetingStates retrieves all meeting states
func GetMeetingStates(
	pool *pgxpool.Pool,
	q sq.SelectBuilder,
) ([]*MeetingState, error) {
	ctx := context.Background()
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
	rows, err := pool.Query(ctx, qry, params...)
	if err != nil {
		return nil, err
	}
	cmd := rows.CommandTag()
	results := make([]*MeetingState, 0, cmd.RowsAffected())
	for rows.Next() {
		state, err := meetingStateFromRow(pool, rows)
		if err != nil {
			return nil, err
		}
		results = append(results, state)
	}
	return results, nil
}

// GetMeetingState tries to retriev a single meeting state
func GetMeetingState(pool *pgxpool.Pool, q sq.SelectBuilder) (*MeetingState, error) {
	states, err := GetMeetingStates(pool, q)
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		return nil, nil
	}
	return states[0], nil
}

func meetingStateFromRow(
	pool *pgxpool.Pool,
	row pgx.Row,
) (*MeetingState, error) {
	state := InitMeetingState(pool, &MeetingState{})
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
func (s *MeetingState) GetBackendState() (*BackendState, error) {
	if s.BackendID == nil {
		return nil, nil
	}

	if s.backend != nil {
		return s.backend, nil
	}

	// Get related backend state
	var err error
	s.backend, err = GetBackendState(s.pool, Q().
		Where("id = ?", s.BackendID))
	if err != nil {
		return nil, err
	}

	return s.backend, nil
}

// GetFrontendState loads the frontend state for the meeting
func (s *MeetingState) GetFrontendState() (*FrontendState, error) {
	if s.FrontendID == nil {
		return nil, nil
	}

	if s.frontend != nil {
		return s.frontend, nil
	}

	// Load frontend state from database
	var err error
	s.frontend, err = GetFrontendState(s.pool, Q().
		Where("id = ?", s.FrontendID))
	if err != nil {
		return nil, err
	}

	return s.frontend, nil
}

// DeleteMeetingStateByID will remove a meeting state.
// It will succeed, even if no such meeting was present.
func DeleteMeetingStateByID(pool *pgxpool.Pool, id string) error {
	ctx := context.Background()
	qry := `
		DELETE FROM meetings WHERE id = $1
	`
	_, err := pool.Exec(ctx, qry, id)
	return err
}

// DeleteMeetingStateByInternalID will remove a meeting state.
// It will succeed, even if no such meeting was present.
func DeleteMeetingStateByInternalID(pool *pgxpool.Pool, id string) error {
	ctx := context.Background()
	qry := `
		DELETE FROM meetings WHERE internal_id = $1
	`
	_, err := pool.Exec(ctx, qry, id)
	return err
}

// Refresh the backend state from the database
func (s *MeetingState) Refresh() error {
	// Load from database
	next, err := GetMeetingState(s.pool, Q().
		Where("id = ?", s.ID))
	if err != nil {
		return err
	}
	*s = *next
	return nil
}

// Save updates or inserts a meeting state into our
// cluster state.
func (s *MeetingState) Save() error {
	var (
		err error
		id  string
	)
	if s.CreatedAt.IsZero() {
		id, err = s.insert()
		s.ID = id
	} else {
		err = s.update()
	}
	if err != nil {
		return err
	}

	return s.Refresh()
}

// Add a new meeting to the database
func (s *MeetingState) insert() (string, error) {
	ctx := context.Background()
	qry := `
		INSERT INTO meetings (
			id,
			internal_id,

			state,

			frontend_id,
			backend_id
		) VALUES (
			$1, $2, $3, $4
		) RETURNING id`
	err := s.pool.QueryRow(ctx, qry,
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
func (s *MeetingState) update() error {
	ctx := context.Background()

	s.UpdatedAt = time.Now().UTC()
	qry := `
		UPDATE meetings
		   SET state		= $2,
		       frontend_id  = $3,
			   backend_id   = $4,
		  	   synced_at    = $5,
			   updated_at   = $6 
	 	 WHERE id = $1`
	_, err := s.pool.Exec(ctx, qry,
		s.ID,
		s.Meeting,
		s.FrontendID,
		s.BackendID,
		s.SyncedAt,
		s.UpdatedAt)
	return err
}

// SetBackendID associates a meeting with a backend
func (s *MeetingState) SetBackendID(id string) error {
	if s.BackendID != nil && *s.BackendID == id {
		return nil // nothing to do here
	}

	// Bind backend
	s.BackendID = &id
	ctx := context.Background()
	qry := `UPDATE meetings SET backend_id = $2
			WHERE id = $1`
	_, err := s.pool.Exec(ctx, qry, s.ID, id)
	return err
}

// IsStale checks if the last sync is longer
// ago than a given threashold.
func (s *MeetingState) IsStale() bool {
	return time.Now().UTC().Sub(s.SyncedAt) > 1*time.Minute
}
