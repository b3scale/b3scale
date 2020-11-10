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
	ID string

	Meeting  *bbb.Meeting
	Frontend *FrontendState
	Backend  *BackendState

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
		"id",
		"frontend_id", "backend_id",
		"state",
		"created_at", "updated_at", "synced_at").
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
func GetMeetingState(conn *pgxpool.Pool, q sq.SelectBuilder) (*MeetingState, error) {
	states, err := GetMeetingStates(conn, q)
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		return nil, nil
	}
	return states[0], nil
}

func meetingStateFromRow(
	conn *pgxpool.Pool,
	row pgx.Row,
) (*MeetingState, error) {
	state := InitMeetingState(conn, &MeetingState{})
	var (
		frontendID string
		backendID  string
	)

	err := row.Scan(
		&state.ID,
		&frontendID,
		&backendID,
		&state.Meeting,
		&state.CreatedAt,
		&state.UpdatedAt,
		&state.SyncedAt)
	if err != nil {
		return nil, err
	}

	// Get related backend state
	state.Backend, err = GetBackendState(conn, Q().
		Where("id = ?", backendID))
	if err != nil {
		return nil, err
	}
	state.Frontend, err = GetFrontendState(conn, Q().
		Where("id = ?", frontendID))
	if err != nil {
		return nil, err
	}

	return state, err
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
	id := s.Meeting.MeetingID
	var (
		frontendID *string
		backendID  *string
	)
	if s.Frontend != nil {
		frontendID = &s.Frontend.ID
	}
	if s.Backend != nil {
		backendID = &s.Backend.ID
	}

	ctx := context.Background()
	qry := `
		INSERT INTO meetings (
			id,
			state,

			frontend_id,
			backend_id
		) VALUES (
			$1, $2, $3, $4
		) RETURNING id`
	err := s.pool.QueryRow(ctx, qry,
		id,
		s.Meeting,
		frontendID,
		backendID).Scan(&s.ID)
	if err != nil {
		return "", err
	}

	return s.ID, nil
}

// Update the meeting state
func (s *MeetingState) update() error {
	ctx := context.Background()

	var (
		frontendID *string
		backendID  *string
	)
	if s.Frontend != nil {
		frontendID = &s.Frontend.ID
	}
	if s.Backend != nil {
		backendID = &s.Backend.ID
	}

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
		frontendID,
		backendID,
		s.SyncedAt,
		s.UpdatedAt)
	return err
}

// IsStale checks if the last sync is longer
// ago than a given threashold.
func (s *MeetingState) IsStale() bool {
	return time.Now().UTC().Sub(s.SyncedAt) > 1*time.Minute

}
