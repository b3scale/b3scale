package cluster

import (
	"errors"
	"fmt"
	"log"
	"time"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// A Backend is a BigBlueButton instance in the cluster.
//
// It has a bbb.backend secret for request authentication,
// stored in the backend state. The state is shared across all
// instances.
//
type Backend struct {
	state  *store.BackendState
	client *bbb.Client
}

// NewBackend creates a new backend instance with
// a fresh bbb client.
func NewBackend(state *store.BackendState) *Backend {
	return &Backend{
		client: bbb.NewClient(),
		state:  state,
	}
}

// Backend State Sync: loadNodeState will make
// a small request to get a meeting that does not
// exist to check if the credentials are valid.
func (b *Backend) loadNodeState() error {
	log.Println(b.state.ID, "SYNC: node state")
	defer b.state.Save()

	// Measure latency
	t0 := time.Now()
	res, err := b.IsMeetingRunning(bbb.IsMeetingRunningRequest(
		bbb.Params{
			"meetingID": "00000000-0000-0000-0000-000000000001",
		}))
	t1 := time.Now()
	latency := t1.Sub(t0)

	if err != nil {
		errMsg := fmt.Sprintf("%s", err)
		b.state.NodeState = "error"
		b.state.LastError = &errMsg
		return errors.New(errMsg)
	}

	if res.Returncode != "SUCCESS" {
		// Update backend state
		errMsg := fmt.Sprintf("%s: %s", res.MessageKey, res.Message)
		b.state.LastError = &errMsg
		b.state.NodeState = "error"
		return errors.New(errMsg)
	}

	// Update state
	b.state.LastError = nil
	b.state.Latency = latency
	b.state.NodeState = "ready"
	return err
}

// BBB API Implementation

// Create a new Meeting
func (b *Backend) Create(req *bbb.Request) (
	*bbb.CreateResponse, error,
) {
	req = req.WithBackend(b.state.Backend)
	// Make request to the backend and update local
	// meetings state
	res, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	createRes := res.(*bbb.CreateResponse)

	// Insert meeting into state
	err = b.state.CreateMeeting(req.Frontend, createRes.Meeting)
	if err != nil {
		return nil, err
	}

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
	req = req.WithBackend(b.state.Backend)
	res, err := b.client.Do(req)

	return res.(*bbb.IsMeetingRunningResponse), err
}

// End a meeting
func (b *Backend) End(req *bbb.Request) (*bbb.EndResponse, error) {
	req = req.WithBackend(b.state.Backend)
	res, err := b.client.Do(req)
	return res.(*bbb.EndResponse), err
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
