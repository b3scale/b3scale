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

// The RIB provides an interface for retrieving
// stored routing decisions for a meeting.
type RIB interface {
	// Routing information:
	// Backend
	GetBackend(string) (*cluster.Backend, error)
	SetBackend(string, *cluster.Backend) error

	// Frontend
	GetFrontend(string) (*cluster.Frontend, error)
	SetFrontend(string, *cluster.Frontend) error

	// Forget about the meeting
	Delete(string) error
}

// The BackendState is shared across instances
// and encapsulates the list of meetings and recordings.
// The backend.ID should be used as identifier.
type BackendState interface {
	// Meetings
	GetMeetings(string) (bbb.MeetingsCollection, error)
	SetMeetings(string, bbb.MeetingsCollection) error

	// Recordings
	GetRecordings(string) (bbb.MeetingsCollection, error)
	SetRecordings(string, bbb.MeetingsCollection) error

	// Forget about the backend
	Delete(string)
}
