package cluster

import (
	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// The RIB provides an interface for retrieving
// stored routing decisions for a meeting.
type RIB struct{}

// Routing information

// GetBackend retrievs the associated backend
// for a meeting.
func (r *RIB) GetBackend(*bbb.Meeting) (*cluster.Backend, error) {
	return nil, nil
}

// SetBackend associates a meeting with a backend
func (r *RIB) SetBackend(*bbb.Meeting, *cluster.Backend) error {
	return nil
}

// GetFrontend retriefs the associated frontend with a
// meeting.
func (r *RIB) GetFrontend(*bbb.Meeting) (*cluster.Frontend, error) {
	return nil, nil
}

// SetFrontend associates a meeting with a frontend
func (r *RIB) SetFrontend(*bbb.Meeting, *cluster.Frontend) error {
	return nil
}

// Delete forgets about the meeting
func (r *RIB) Delete(m *bbb.Meeting) error {
	return nil
}
