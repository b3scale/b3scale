package cluster

import (
	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The RIB provides an interface for retrieving
// stored routing decisions for a meeting.
type RIB struct {
	conn *pgxpool.Pool
}

// NewRIB creates a new routing information base
func NewRIB(conn *pgxpool.Pool) *RIB {
	return &RIB{
		conn: conn,
	}
}

// GetBackend retrievs the associated backend
// for a meeting.
func (r *RIB) GetBackend(*bbb.Meeting) (*Backend, error) {
	return nil, nil
}

// SetBackend associates a meeting with a backend
func (r *RIB) SetBackend(*bbb.Meeting, *Backend) error {
	return nil
}

// GetFrontend retriefs the associated frontend with a
// meeting.
func (r *RIB) GetFrontend(*bbb.Meeting) (*Frontend, error) {
	return nil, nil
}

// SetFrontend associates a meeting with a frontend
func (r *RIB) SetFrontend(*bbb.Meeting, *Frontend) error {
	return nil
}

// Delete forgets about the meeting
func (r *RIB) Delete(m *bbb.Meeting) error {
	return nil
}
