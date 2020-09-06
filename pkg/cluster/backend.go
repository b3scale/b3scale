package cluster

import (
	"fmt"
	"log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

const (
	// BackendStateInit when syncing the state
	BackendStateInit = iota
	// BackendStateReady when we accept requests
	BackendStateReady
	// BackendStateError when we do not accept requests
	BackendStateError
)

// The BackendState can be init, active or error
type BackendState int

// A Backend is a BigBlueButton instance and a node in
// the cluster.
//
// It has a host and a secret for request authentication.
// It syncs it's state with the bbb instance.
type Backend struct {
	ID        string
	State     BackendState
	LastError string

	config *config.Backend
	client *bbb.Client
}

// NewBackend creates a cluster node.
func NewBackend(config *config.Backend) *Backend {
	return &Backend{
		ID:     config.Host,
		config: config,
		client: nil,
	}
}

// Start the backend
func (b *Backend) Start() {
	log.Println("Starting backend:", b.ID)
}

// Stop shuts down the backend process
func (b *Backend) Stop() {
	log.Println("Shutting down backend:", b.ID)
}

// BBB API Implementation
// TODO: Fix Response Types

// Create a new Meeting
func (b *Backend) Create(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}

// Join a meeting
func (b *Backend) Join(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}

// IsMeetingRunning returns the is meeting running state
func (b *Backend) IsMeetingRunning(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}

// End a meeting
func (b *Backend) End(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}

// GetMeetingInfo gets the meeting details
func (b *Backend) GetMeetingInfo(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}

// GetMeetings retrieves a list of meetings
func (b *Backend) GetMeetings(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}

// GetRecordings retrieves a list of recordings
func (b *Backend) GetRecordings(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}

// PublishRecordings publishes a recording
func (b *Backend) PublishRecordings(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}

// DeleteRecordings deletes recordings
func (b *Backend) DeleteRecordings(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}

// UpdateRecordings updates recordings
func (b *Backend) UpdateRecordings(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}

// GetDefaultConfigXML retrieves the default config xml
func (b *Backend) GetDefaultConfigXML(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}

// SetConfigXML sets the? config xml
func (b *Backend) SetConfigXML(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}

// GetRecordingTextTracks retrievs all text tracks
func (b *Backend) GetRecordingTextTracks(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}

// PutRecordingTextTrack adds a text track
func (b *Backend) PutRecordingTextTrack(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}
