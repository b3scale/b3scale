package cluster

import (
	"fmt"
	"log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// A Backend is a BigBlueButton instance in the cluster.
//
// It has a bbb.backend secret for request authentication,
// stored in the backend state. The state is shared across all
// instances.
//
type Backend struct {
	state  *BackendState
	client *bbb.Client
}

// Load current state from the node. This includes
// all meetings, meetings in detail and recordings.
func (b *Backend) fetchBBBState() error {
	if err := b.fetchMeetingsState(); err != nil {
		return err
	}

	return nil
}

// Backend State Sync: Loads the state from
// the bbb backend and keeps it locally.
// Meeting details will be fetched.
func (b *Backend) fetchMeetingsState() error {
	log.Println(b.state.ID, "SYNC: meetings")
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

	if err := b.state.SetMeetings(stateMeetings); err != nil {
		return err
	}

	log.Println(b.state.ID, "Meetings:", stateMeetings)

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
	if err := b.state.AddMeeting(createRes.Meeting); err != nil {
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
