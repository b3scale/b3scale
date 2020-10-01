package cluster

/*
 The RIB provides information for routing
 decisions: You can look up the stored frontend or
 backend for a given meeting id.

 This state needs to be shared across instances
 and must be persisted.
*/

import (
	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// RIB provides an interface for retrieving
// routing information.
type RIB interface {
	// Backend
	GetBackend(*bbb.Meeting) (*Backend, error)
	SetBackend(*bbb.Meeting, *Backend) error

	// Frontend
	GetFrontend(*bbb.Meeting) (*Frontend, error)
	SetFrontend(*bbb.Meeting, *Backend) error

	// Meeting
	Delete(*bbb.Meeting) error
	Meetings() ([]*bbb.Meeting, error)
}
