package store

/*
 The Store provides information for routing
 decisions: You can look up the stored frontend or
 backend for a given meeting id.

 This state needs to be shared across instances
 and must be persisted.
*/

import (
	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// Store provides an interface for persisting and
// retrieving data from a shared state. The store is
// stringly typed.
type Store interface {
	Get(string) (string, error)
	GetAll(string) ([]string, error)

	// Set a key with data
	Set(string, string) error

	// Delete a key
	Delete(string) error
}

// The RIB provides an interface for retrieving
// stored routing decisions for a meeting.
type RIB interface {
	// Routing information:
	// Backend
	GetBackend(*bbb.Meeting) (*cluster.Backend, error)
	SetBackend(*bbb.Meeting, *cluster.Backend) error

	// Frontend
	GetFrontend(*bbb.Meeting) (*cluster.Frontend, error)
	SetFrontend(*bbb.Meeting, *cluster.Frontend) error

	// Forget about the meeting
	Delete(*bbb.Meeting) error
}

// The BackendState is shared across instances
// and encapsulates the list of meetings and recordings.
// The backend.ID should be used as identifier.
type BackendState interface {
	// Meetings
	GetMeetings(*cluster.Backend) (bbb.MeetingsCollection, error)
	SetMeetings(*cluster.Backend, bbb.MeetingsCollection) error

	// Recordings
	GetRecordings(*cluster.Backend) (bbb.MeetingsCollection, error)
	SetRecordings(*cluster.Backend, bbb.MeetingsCollection) error

	// Forget about the backend
	Delete(*cluster.Backend)
}
