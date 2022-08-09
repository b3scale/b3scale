package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"

	"github.com/b3scale/b3scale/pkg/bbb"
)

// Errors
var (
	ErrFrontendRequired = errors.New("meeting requires a frontend state")
)

// The BackendState is shared across b3scale instances
// and encapsulates the list of meetings and recordings.
// The backend.ID should be used as identifier.
type BackendState struct {
	ID string `json:"id"`

	NodeState  string `json:"node_state"`
	AdminState string `json:"admin_state"`

	AgentHeartbeat time.Time `json:"agent_heartbeat"`
	AgentRef       *string   `json:"agent_ref"`

	LastError *string `json:"last_error"`

	Latency        time.Duration `json:"latency"`
	MeetingsCount  uint          `json:"meetings_count"`
	AttendeesCount uint          `json:"attendees_count"`

	LoadFactor float64 `json:"load_factor"`

	Backend *bbb.Backend `json:"bbb"`

	Settings BackendSettings `json:"settings"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	SyncedAt  time.Time `json:"synced_at"`
}

// AgentHeartbeat is a short api response
type AgentHeartbeat struct {
	BackendID string    `json:"backend_id"`
	Heartbeat time.Time `json:"heartbeat"`
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
	if init.LoadFactor == 0 {
		init.LoadFactor = 1.0
	}
	return init
}

// GetBackendStates retrievs all backends
func GetBackendStates(
	ctx context.Context,
	tx pgx.Tx,
	q sq.SelectBuilder,
) ([]*BackendState, error) {
	qry, params, _ := q.From("backends").Columns(
		"backends.id",
		"backends.node_state",
		"backends.admin_state",
		"backends.agent_heartbeat",
		"backends.agent_ref",
		"backends.last_error",
		"backends.latency",
		"backends.meetings_count",
		"backends.attendees_count",
		"backends.load_factor",
		"backends.host",
		"backends.secret",
		"backends.settings",
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
		state := InitBackendState(&BackendState{})
		err := rows.Scan(
			&state.ID,
			&state.NodeState,
			&state.AdminState,
			&state.AgentHeartbeat,
			&state.AgentRef,
			&state.LastError,
			&state.Latency,
			&state.MeetingsCount,
			&state.AttendeesCount,
			&state.LoadFactor,
			&state.Backend.Host,
			&state.Backend.Secret,
			&state.Settings,
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
	ctx context.Context,
	tx pgx.Tx,
	q sq.SelectBuilder,
) (*BackendState, error) {
	states, err := GetBackendStates(ctx, tx, q)
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		return nil, nil
	}
	return states[0], nil
}

// Refresh the backend state from the database
func (s *BackendState) Refresh(
	ctx context.Context,
	tx pgx.Tx,
) error {
	// Load from database
	next, err := GetBackendState(ctx, tx, Q().Where(
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
func (s *BackendState) Save(
	ctx context.Context,
	tx pgx.Tx,
) error {
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

	return s.Refresh(ctx, tx)
}

// Private insert: adds a new row to the backends table
func (s *BackendState) insert(
	ctx context.Context,
	tx pgx.Tx,
) (string, error) {
	qry := `
		INSERT INTO backends (
			host,
			secret,

			node_state,
			admin_state,

			settings,

			load_factor,

			agent_ref
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	insertID := ""
	err := tx.QueryRow(ctx, qry,
		// Values
		s.Backend.Host,
		s.Backend.Secret,
		s.NodeState,
		s.AdminState,
		s.Settings,
		s.LoadFactor,
		s.AgentRef).Scan(&insertID)

	return insertID, err
}

// Private update: updates the db row
func (s *BackendState) update(
	ctx context.Context,
	tx pgx.Tx,
) error {
	qry := `
		UPDATE backends
		   SET node_state   = $2,
		       admin_state  = $3,

			   last_error   = $4,

			   latency      = $5,

			   host         = $6,
			   secret       = $7,

			   settings     = $8,

			   load_factor  = $9,

			   synced_at    = $10,
			   updated_at   = $11

		 WHERE id = $1
	`
	_, err := tx.Exec(
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
		s.Settings,
		s.LoadFactor,
		s.SyncedAt,
		time.Now().UTC())

	return err
}

// Delete will remove the backend from the store
func (s *BackendState) Delete(
	ctx context.Context,
	tx pgx.Tx,
) error {
	// For now we take all the meetings with us.
	qry := `
		DELETE FROM meetings WHERE backend_id = $1
	`
	if _, err := tx.Exec(ctx, qry, s.ID); err != nil {
		return err
	}

	qry = `
		DELETE FROM backends WHERE id = $1
	`
	if _, err := tx.Exec(ctx, qry, s.ID); err != nil {
		return err
	}

	return nil
}

// UpdateAgentHeartbeat will set the attribute to the
// current timestamp
func (s *BackendState) UpdateAgentHeartbeat(
	ctx context.Context,
	tx pgx.Tx,
) (*AgentHeartbeat, error) {
	qry := `
		UPDATE backends
		   SET agent_heartbeat = $2
		 WHERE id = $1
	`

	now := time.Now().UTC()
	_, err := tx.Exec(ctx, qry, s.ID, now)
	if err != nil {
		return nil, err
	}
	s.AgentHeartbeat = now

	heartbeat := &AgentHeartbeat{
		BackendID: s.ID,
		Heartbeat: s.AgentHeartbeat,
	}
	return heartbeat, nil
}

// IsAgentAlive checks if the heartbeat is older
// than the threshold
func (s *BackendState) IsAgentAlive() bool {
	threshold := 5 * time.Second
	now := time.Now().UTC()
	return now.Sub(s.AgentHeartbeat) <= threshold
}

// IsNodeReady checks if the agent is alive and the node
// state is ready
func (s *BackendState) IsNodeReady() bool {
	return s.IsAgentAlive() && s.NodeState == "ready"
}

// ClearMeetings will remove all meetings in the current state
func (s *BackendState) ClearMeetings(
	ctx context.Context,
	tx pgx.Tx,
) error {
	qry := `
		DELETE FROM meetings WHERE backend_id = $1
	`
	_, err := tx.Exec(ctx, qry, s.ID)
	if err != nil {
		return err
	}

	return err
}

// Internal: updateBackendStatCounters counts meetings
// and attendees for a given backendID
func updateBackendStatCounters(
	ctx context.Context,
	tx pgx.Tx,
	backendID string,
) error {
	// Get meeting states and refresh counters
	mstates, err := GetMeetingStates(ctx, tx, Q().
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

	qry := `
		UPDATE backends
		   SET meetings_count = $2,
		       attendees_count = $3
		 WHERE backends.id = $1
	`
	if _, err := tx.Exec(ctx, qry, backendID, mcount, acount); err != nil {
		return err
	}

	return nil
}

// UpdateStatCounters counts meetings and attendees and updates the properties
func (s *BackendState) UpdateStatCounters(
	ctx context.Context,
	tx pgx.Tx,
) error {
	return updateBackendStatCounters(ctx, tx, s.ID)
}

// CreateMeetingState will create a new state for the
// current backend state. A frontend is attached if present.
func (s *BackendState) CreateMeetingState(
	ctx context.Context,
	tx pgx.Tx,
	frontend *bbb.Frontend,
	meeting *bbb.Meeting,
) (*MeetingState, error) {
	mstate := InitMeetingState(&MeetingState{
		BackendID: &s.ID,
		Meeting:   meeting,
	})
	mstate.MarkSynced()

	// Attach frontend if present
	if frontend != nil {
		// Combine frontend and backend state together
		// with meeting data into a meeting state.
		fstate, err := GetFrontendState(ctx, tx, Q().
			Where("key = ?", frontend.Key))
		if err != nil {
			return nil, err
		}
		mstate.FrontendID = &fstate.ID
	}

	if err := mstate.Save(ctx, tx); err != nil {
		return nil, err
	}

	return mstate, nil
}

// CreateOrUpdateMeetingState will try to update the meeting
// or will create a new meeting state if the meeting does
// not exists. The new meeting will not be associated with
// a frontend state - however the meeting can later be claimed
// by a frontend.
func (s *BackendState) CreateOrUpdateMeetingState(
	ctx context.Context,
	tx pgx.Tx,
	meeting *bbb.Meeting,
) error {
	mstate := InitMeetingState(&MeetingState{
		BackendID: &s.ID,
		Meeting:   meeting,
	})
	mstate.MarkSynced()
	if _, err := mstate.Upsert(ctx, tx); err != nil {
		return err
	}
	return nil
}

// Validate the backend state
func (s *BackendState) Validate() ValidationError {
	err := ValidationError{}

	if s.Backend == nil {
		err.Add("bbb", "this field is required")
		return err
	}

	// BBB hostname
	host := s.Backend.Host
	if host == "" {
		err.Add("bbb.host", ErrFieldRequired)
	}

	if !strings.HasPrefix(host, "http") {
		err.Add("bbb.host", "should start with http(s)://")
	}
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}

	// Secret
	secret := strings.TrimSpace(s.Backend.Secret)
	if secret == "" {
		err.Add("bbb.secret", ErrFieldRequired)
	}

	if len(err) > 0 {
		return err
	}

	return nil
}
