package store

import (
	"context"
	//	"fmt"
	"time"

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
	UpdatedAt *time.Time
	SyncedAt  *time.Time

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

// GetMeetingState retrieves a meeting state
func GetMeetingState(
	pool *pgxpool.Pool,
	q *Query,
) (*MeetingState, error) {
	ctx := context.Background()
	qry := `
		SELECT
		  id,

		  frontend_id,
		  backend_id,

		  state,

		  created_at,
		  updated_at,
		  synced_at
		FROM  meetings ` + q.related() + `
		WHERE ` + q.where()
}

// Refresh the backend state from the database
func (s *MeetingState) Refresh() error {
	// Load from database
	next, err := GetMeetingState(s.conn, NewQuery().
		Eq("id", s.ID))
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
