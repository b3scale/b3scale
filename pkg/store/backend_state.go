package store

import (
	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// The BackendState is shared across b3scale instances
// and encapsulates the list of meetings and recordings.
// The backend.ID should be used as identifier.
type BackendState struct {
}

// Meetings
func (s *BackendState) GetMeetings() (bbb.MeetingsCollection, error) {
	return nil, nil
}

// SetMeetings
func (s *BackendState) SetMeetings(bbb.MeetingsCollection) error {

}

/*

	// Recordings
	GetRecordings(*cluster.Backend) (bbb.RecordingsCollection, error)
	SetRecordings(*cluster.Backend, bbb.RecordingsCollection) error

	// Forget about the backend
	Delete(*cluster.Backend)

*/
