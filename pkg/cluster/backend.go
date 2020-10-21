package cluster

import (
	"fmt"
	"log"
	"time"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
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
	ID string

	cfg    *config.Backend
	state  *store.BackendState
	client *bbb.Client

	stop chan bool
}

// NewBackend creates a cluster node.
func NewBackend(
	cfg *config.Backend,
	state *store.BackendState,
) *Backend {
	// Start HTTP client
	client := bbb.NewClient(cfg)
	return &Backend{
		ID:     cfg.Host,
		cfg:    cfg,
		client: client,
		state:  state,
		stop:   make(chan bool),
	}
}

// Start the backend
func (b *Backend) Start() {
	log.Println("Starting backend:", b.ID)
	// Main Loop
	for {
		select {
		case <-b.stop:
			b.shutdown()
		default:
			b.loop()
		}
	}
}

// Backend Main
func (b *Backend) loop() {
	// Initial sync
	err := b.loadNodeState()
	if err != nil {
		b.State = BackendStateReady
		return
	}

	b.State = BackendStateReady

	// Wait.
	time.Sleep(10000 * time.Hour)
}

// Beckend on shutdown
func (b *Backend) shutdown() {
	log.Println("Shutting down backend:", b.ID)
}

// Stop shuts down the backend process
func (b *Backend) Stop() {
	b.stop <- true
}

// Load current state from the node. This includes
// all meetings, meetings in detail and recordings.
func (b *Backend) loadNodeState() error {
	if err := b.loadMeetingsState(); err != nil {
		return err
	}

	return nil
}

// Backend State Sync: Loads the state from
// the bbb backend and keeps it locally.
// Meeting details will be fetched.
func (b *Backend) loadMeetingsState() error {
	log.Println(b.ID, "SYNC: meetings")
	// Fetch meetings from backend
	req := bbb.GetMeetingsRequest(bbb.Params{})
	res, err := b.client.Do(req)
	if err != nil {
		return err
	}
	meetingsRes := res.(*bbb.GetMeetingsResponse)

	// Get meeting details
	meetings := meetingsRes.Meetings
	stateMeetings := make([]*bbb.Meeting, 0, len(meetings))
	for _, m := range meetings {
		req = bbb.GetMeetingInfoRequest(bbb.Params{
			bbb.ParamMeetingID: m.MeetingID,
		})
		res, err = b.client.Do(req)
		if err != nil {
			return err // Sync must be complete.
		}
		meetingRes := res.(*bbb.GetMeetingInfoResponse)
		stateMeetings = append(stateMeetings, meetingRes.Meeting)
	}

	b.meetings = stateMeetings
	log.Println(b.ID, "Meetings:", b.meetings)

	return nil
}

// BBB API Implementation

// Create a new Meeting
func (b *Backend) Create(req *bbb.Request) (
	*bbb.CreateResponse, error,
) {
	// Make request to the backend and update local
	// meetings state
	res, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	createRes := res.(*bbb.CreateResponse)

	// Insert meeting into state
	b.meetings = append(b.meetings, createRes.Meeting)

	return createRes, nil
}

// Join a meeting
func (b *Backend) Join(
	req *bbb.Request,
) (*bbb.JoinResponse, error) {
	// Make join request to the backend and update local
	// meetings state
	res, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	joinRes := res.(*bbb.JoinResponse)

	// Update meeting attendee list
	// ...

	return joinRes, nil
}

// IsMeetingRunning returns the is meeting running state
func (b *Backend) IsMeetingRunning(
	req *bbb.Request,
) (*bbb.IsMeetingRunningResponse, error) {

	// Try With the server...

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
func (b *Backend) DeleteRecordings(
	req *bbb.Request,
) (*bbb.DeleteRecordingsResponse, error) {
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
