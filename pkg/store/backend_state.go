package store

import (
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The BackendState is shared across b3scale instances
// and encapsulates the list of meetings and recordings.
// The backend.ID should be used as identifier.
type BackendState struct {
	ID string

	NodeState  string
	AdminState string

	LastError string

	Backend *bbb.Backend

	Tags []string

	CreatedAt time.Time
	UpdatedAt time.Time
	SyncedAt  time.Time

	// DB
	conn *pgxpool.Pool
}

// InitBackendState initializes a new backend state with
// an initial state.
func InitBackendState(conn *pgxpool.Pool, initial *BackendState) *BackendState {
	initial.conn = conn
	return initial
}

// Save persists the backend state in the database store
func (s *BackendState) Save() error {
	return nil
}

// GetMeetings retrievs all meetings for a meeting
// filterable with GetMeetingsOpts.
func (s *BackendState) GetMeetings() (bbb.MeetingsCollection, error) {
	return nil, nil
}

// AddMeeting persists a meeting in the store
func (s *BackendState) AddMeeting(fe *bbb.Frontend, m *bbb.Meeting) error {
	return nil
}

/*

	// Recordings
	GetRecordings(*cluster.Backend) (bbb.RecordingsCollection, error)
	SetRecordings(*cluster.Backend, bbb.RecordingsCollection) error

	// Forget about the backend
	Delete(*cluster.Backend)

*/
