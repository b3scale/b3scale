package rib

import (
	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

/*
 The RIB provides information for routing
 decisions: You can look up the stored frontend or
 backend for a given meeting id.

 This state needs to be shared across instances
 and must be persisted.

 We are using redis for this.
*/

// Store provides an interface for retrieving
// routing information.
type Store interface {
	// Backend
	GetBackend(*bbb.Meeting) (*cluster.Backend, error)
	SetBackend(*bbb.Meeting, *cluster.Backend) error

	// Frontend
	GetFrontend(*bbb.Meeting) (*cluster.Frontend, error)
	SetFrontend(*bbb.Meeting, *cluster.Backend) error

	// Meeting
	Delete(*bbb.Meeting) error
}
