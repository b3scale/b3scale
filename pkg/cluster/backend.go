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
func (b *Backend) Create(req *bbb.Request) (
	*bbb.CreateResponse, error,
) {
	return nil, fmt.Errorf("implement me")
}

// Join a meeting
func (b *Backend) Join(
	req *bbb.Request,
) (*bbb.JoinResponse, error) {
	return nil, fmt.Errorf("implement me")
}

// IsMeetingRunning returns the is meeting running state
func (b *Backend) IsMeetingRunning(
	req *bbb.Request,
) (*bbb.IsMeetingRunningResponse, error) {
	return nil, fmt.Errorf("implement me")
}

// End a meeting
func (b *Backend) End(req *bbb.Request) (*bbb.EndResponse, error) {
	return nil, fmt.Errorf("implement me")
}

// GetMeetingInfo gets the meeting details
func (b *Backend) GetMeetingInfo(
	req *bbb.Request,
) (*bbb.GetMeetingInfoResponse, error) {
	return nil, fmt.Errorf("implement me")
}

// GetMeetings retrieves a list of meetings
func (b *Backend) GetMeetings(
	req *bbb.Request,
) (*bbb.GetMeetingsResponse, error) {
	return nil, fmt.Errorf("implement me")
}

// GetRecordings retrieves a list of recordings
func (b *Backend) GetRecordings(
	req *bbb.Request,
) (*bbb.GetRecordingsResponse, error) {
	return nil, fmt.Errorf("implement me")
}

// PublishRecordings publishes a recording
func (b *Backend) PublishRecordings(
	req *bbb.Request,
) (*bbb.PublishRecordingsResponse, error) {
	return nil, fmt.Errorf("implement me")
}

// DeleteRecordings deletes recordings
func (b *Backend) DeleteRecordings(req *Request) (*Response, error) {
	return nil, fmt.Errorf("implement me")
}

// UpdateRecordings updates recordings
func (b *Backend) UpdateRecordings(
	req *bbb.Request,
) (*bbb.UpdateRecordingsResponse, error) {
	return nil, fmt.Errorf("implement me")
}

// GetDefaultConfigXML retrieves the default config xml
func (b *Backend) GetDefaultConfigXML(
	req *bbb.Request,
) (*bbb.GetDefaultConfigXMLResponse, error) {
	return nil, fmt.Errorf("implement me")
}

// SetConfigXML sets the? config xml
func (b *Backend) SetConfigXML(
	req *bbb.Request,
) (*bbb.SetConfigXMLResponse, error) {
	return nil, fmt.Errorf("implement me")
}

// GetRecordingTextTracks retrievs all text tracks
func (b *Backend) GetRecordingTextTracks(
	req *bbb.Request,
) (*bbb.GetRecordingTextTracksResponse, error) {
	return nil, fmt.Errorf("implement me")
}

// PutRecordingTextTrack adds a text track
func (b *Backend) PutRecordingTextTrack(
	req *bbb.Request,
) (*bbb.PutRecordingTextTrackResponse, error) {
	return nil, fmt.Errorf("implement me")
}
