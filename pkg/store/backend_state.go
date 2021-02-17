package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
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

	AgentHeartbeat time.Time

	LastError *string

	Latency        time.Duration
	MeetingsCount  uint
	AttendeesCount uint

	LoadFactor float64

	Backend *bbb.Backend

	Tags []string

	CreatedAt time.Time
	UpdatedAt time.Time
	SyncedAt  time.Time
}

// InitBackendState initializes a new backend state with
// an initial state.
func InitBackendState(init *BackendState) *BackendState {
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
	return init
}

// GetBackendStates retrievs all backends
func GetBackendStates(
	pool *pgxpool.Pool,
	q sq.SelectBuilder,
) ([]*BackendState, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second)
	defer cancel()

	// To utilize the locking of the we wrap this in a transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qry, params, _ := q.From("backends").Columns(
		"backends.id",
		"backends.node_state",
		"backends.admin_state",
		"backends.agent_heartbeat",
		"backends.last_error",
		"backends.latency",
		"backends.meetings_count",
		"backends.attendees_count",
		"backends.load_factor",
		"backends.host",
		"backends.secret",
		"backends.tags",
		"backends.created_at",
		"backends.updated_at",
		"backends.synced_at").
		ToSql()
	// log.Println("SQL:", qry, params)
	rows, err := tx.Query(ctx, qry, params...)
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
			&state.AgentHeartbeat,
			&state.LastError,
			&state.Latency,
			&state.MeetingsCount,
			&state.AttendeesCount,
			&state.LoadFactor,
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
func GetBackendState(
	pool *pgxpool.Pool,
	q sq.SelectBuilder,
) (*BackendState, error) {
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
	next, err := GetBackendState(s.pool, Q().Where(
		sq.Eq{"id": s.ID},
	))
	if err != nil {
		return err
	}
	if next == nil {
		return fmt.Errorf("backend %s is gone", s.ID)
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
	ctx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second)
	defer cancel()

	qry := `
		INSERT INTO backends (
			host,
			secret,

			node_state,
			admin_state,

			tags,

			load_factor
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	insertID := ""
	err := s.pool.QueryRow(ctx, qry,
		// Values
		s.Backend.Host,
		s.Backend.Secret,
		s.NodeState,
		s.AdminState,
		s.Tags, s.LoadFactor).Scan(&insertID)

	return insertID, err
}

// Private update: updates the db row
func (s *BackendState) update() error {
	ctx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second)
	defer cancel()

	qry := `
		UPDATE backends
		   SET node_state   = $2,
		       admin_state  = $3,

			   last_error   = $4,

			   latency      = $5,

			   host         = $6,
			   secret       = $7,

			   tags         = $8,

			   load_factor  = $9,

			   synced_at    = $10,
			   updated_at   = $11

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
		s.LoadFactor,
		s.SyncedAt,
		time.Now().UTC())

	return err
}

// Delete will remove the backend from the store
func (s *BackendState) Delete() error {
	ctx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// For now we take all the meetings with us.
	qry := `
		DELETE FROM meetings WHERE backend_id = $1
	`
	_, err = tx.Exec(ctx, qry, s.ID)
	if err != nil {
		return err
	}

	qry = `
		DELETE FROM backends WHERE id = $1
	`
	_, err = tx.Exec(ctx, qry, s.ID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// UpdateAgentHeartbeat will set the attribute to the
// current timestamp
func (s *BackendState) UpdateAgentHeartbeat() error {
	ctx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second)
	defer cancel()
	qry := `
		UPDATE backends
		   SET agent_heartbeat = $2
		 WHERE id = $1
	`

	now := time.Now().UTC()
	_, err := s.pool.Exec(ctx, qry, s.ID, now)
	if err != nil {
		return err
	}

	s.AgentHeartbeat = now

	return nil
}

// IsAgentAlive checks if the heartbeat is older
// than the threshold
func (s *BackendState) IsAgentAlive() bool {
	threshold := 1 * time.Second
	now := time.Now().UTC()
	return now.Sub(s.AgentHeartbeat) <= threshold
}

// ClearMeetings will remove all meetings in the current state
func (s *BackendState) ClearMeetings() error {
	ctx, cancel := context.WithTimeout(
		context.Background(), 5*time.Second)
	defer cancel()
	qry := `
		DELETE FROM meetings WHERE backend_id = $1
	`
	_, err := s.pool.Exec(ctx, qry, s.ID)
	if err != nil {
		return err
	}

	return err
}

// Internal: updateBackendStatCounters counts meetings
// and attendees for a given backendID
func updateBackendStatCounters(
	pool *pgxpool.Pool,
	backendID string,
) error {
	// Get meeting states and refresh counters
	mstates, err := GetMeetingStates(pool, Q().
		Where("meetings.backend_id = ?", backendID))
	if err != nil {
		return err
	}

	// Meeting and attendees counter
	mcount := len(mstates)
	acount := 0
	for _, m := range mstates {
		acount += len(m.Meeting.Attendees)
	}

	ctx := context.Background()
	qry := `
		UPDATE backends
		   SET meetings_count = $2,
		       attendees_count = $3
		 WHERE backends.id = $1
	`
	if _, err := pool.Exec(ctx, qry, backendID, mcount, acount); err != nil {
		return err
	}

	return nil
}

// UpdateStatCounters counts meetings and attendees and updates the properties
func (s *BackendState) UpdateStatCounters() error {
	return updateBackendStatCounters(s.pool, s.ID)
}

// CreateMeetingState will create a new state for the
// current backend state. A frontend is attached.
func (s *BackendState) CreateMeetingState(
	frontend *bbb.Frontend,
	meeting *bbb.Meeting,
) (*MeetingState, error) {
	// Combine frontend and backend state together
	// with meeting data into a meeting state.
	fstate, err := GetFrontendState(s.pool, Q().
		Where("key = ?", frontend.Key))
	if err != nil {
		return nil, err
	}
	if fstate == nil {
		return nil, ErrFrontendRequired
	}
	mstate := InitMeetingState(s.pool, &MeetingState{
		BackendID:  &s.ID,
		FrontendID: &fstate.ID,
		Meeting:    meeting,
	})
	if err := mstate.Save(); err != nil {
		return nil, err
	}
	return mstate, nil
}
