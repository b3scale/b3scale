package store

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
