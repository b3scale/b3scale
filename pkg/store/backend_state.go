package store

import (
	"context"
	//	"fmt"
	"errors"
	"time"

	// "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// Errors
var (
	ErrFrontendRequired = errors.New("meeting requires a frontend state")
)

// The BackendState is shared across b3scale instances
// and encapsulates the list of meetings and recordings.
// The backend.ID should be used as identifier.
type BackendState struct {
	ID string

	NodeState  string
	AdminState string

	LastError *string

	Latency time.Duration
	Load    float64

	Backend *bbb.Backend

	Tags []string

	CreatedAt time.Time
	UpdatedAt *time.Time
	SyncedAt  *time.Time

	// DB
	pool *pgxpool.Pool
}

// InitBackendState initializes a new backend state with
// an initial state.
func InitBackendState(pool *pgxpool.Pool, init *BackendState) *BackendState {
	// Add default values
	if init.NodeState == "" {
		init.NodeState = "init"
	}
	if init.AdminState == "" {
		init.AdminState = "ready"
	}
	if init.Backend == nil {
		init.Backend = &bbb.Backend{}
	}
	if init.Tags == nil {
		init.Tags = []string{}
	}
	init.pool = pool
	return init
}

// GetBackendStates retrievs all backends
func GetBackendStates(pool *pgxpool.Pool, q *Query) ([]*BackendState, error) {
	ctx := context.Background()
	qry := `
		SELECT
		  B.id,

		  B.node_state,
		  B.admin_state,

		  B.last_error,

		  B.latency,

		  B.host,
		  B.secret,

		  B.tags,

		  B.created_at,
		  B.updated_at,
		  B.synced_at
		FROM backends AS B ` + q.related() + `
		WHERE ` + q.where()
	rows, err := pool.Query(ctx, qry, q.params()...)
	if err != nil {
		return nil, err
	}
	cmd := rows.CommandTag()
	// fmt.Println("Affected rows:", cmd.RowsAffected())
	results := make([]*BackendState, 0, cmd.RowsAffected())
	for rows.Next() {
		state := InitBackendState(pool, &BackendState{})
		err := rows.Scan(
			&state.ID,
			&state.NodeState,
			&state.AdminState,
			&state.LastError,
			&state.Latency,
			&state.Backend.Host,
			&state.Backend.Secret,
			&state.Tags,
			&state.CreatedAt,
			&state.UpdatedAt,
			&state.SyncedAt)
		if err != nil {
			return nil, err
		}
		results = append(results, state)
	}

	return results, nil
}

// GetBackendState tries to retriev a single backend state
func GetBackendState(pool *pgxpool.Pool, q *Query) (*BackendState, error) {
	states, err := GetBackendStates(pool, q)
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		return nil, nil
	}
	return states[0], nil
}

// Refresh the backend state from the database
func (s *BackendState) Refresh() error {
	// Load from database
	q := NewQuery().Eq("id", s.ID)
	next, err := GetBackendState(s.pool, q)
	if err != nil {
		return err
	}
	*s = *next
	return nil
}

// Save persists the backend state in the database store
func (s *BackendState) Save() error {
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

// Private insert: adds a new row to the backends table
func (s *BackendState) insert() (string, error) {
	ctx := context.Background()
	qry := `
		INSERT INTO backends (
			host,
			secret,

			node_state,
			admin_state,

			tags
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	insertID := ""
	err := s.pool.QueryRow(ctx, qry,
		// Values
		s.Backend.Host,
		s.Backend.Secret,
		s.NodeState,
		s.AdminState,
		s.Tags).Scan(&insertID)

	return insertID, err
}

// Private update: updates the db row
func (s *BackendState) update() error {
	now := time.Now().UTC()
	s.UpdatedAt = &now
	ctx := context.Background()
	qry := `
		UPDATE backends
		   SET node_state   = $2,
		       admin_state  = $3,

			   last_error   = $4,

			   latency      = $5,

			   host         = $6,
			   secret       = $7,

			   tags         = $8,

			   updated_at   = $9,
			   synced_at    = $10

		 WHERE id = $1
	`
	_, err := s.pool.Exec(
		ctx, qry,
		// Identifier
		s.ID,
		// Update Values
		s.NodeState,
		s.AdminState,
		s.LastError,
		s.Latency,
		s.Backend.Host,
		s.Backend.Secret,
		s.Tags,
		s.UpdatedAt,
		s.SyncedAt)

	return err
}

// ClearMeetings will remove all meetings in the current state
func (s *BackendState) ClearMeetings() error {
	ctx := context.Background()
	qry := `
		DELETE FROM meetings WHERE backend_id = $1
	`
	_, err := s.pool.Exec(ctx, qry, s.ID)
	return err
}

// CreateMeetingState will create a new state for the
// current backend state. A frontend is attached.
func (s *BackendState) CreateMeetingState(
	frontend *bbb.Frontend,
	meeting *bbb.Meeting,
) (*MeetingState, error) {
	// Combine frontend and backend state together
	// with meeting data into a meeting state.
	fstate, err := GetFrontendState(s.pool, NewQuery().
		Eq("key", frontend.Key))
	if err != nil {
		return nil, err
	}
	if fstate == nil {
		return nil, ErrFrontendRequired
	}
	mstate := InitMeetingState(s.pool, &MeetingState{
		Backend:  s,
		Frontend: fstate,
		Meeting:  meeting,
	})
	if err := mstate.Save(); err != nil {
		return nil, err
	}
	return mstate, nil
}
