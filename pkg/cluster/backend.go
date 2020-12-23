package cluster

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// BackendStates: The state of the cluster backend node.
const (
	BackendStateInit           = "init"
	BackendStateReady          = "ready"
	BackendStateError          = "error"
	BackendStateStopped        = "stopped"
	BackendStateDecommissioned = "decommissioned"
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
	pool   *pgxpool.Pool
	cmds   *store.CommandQueue
}

// NewBackend creates a new backend instance with
// a fresh bbb client.
func NewBackend(pool *pgxpool.Pool, state *store.BackendState) *Backend {
	return &Backend{
		client: bbb.NewClient(),
		state:  state,
		pool:   pool,
		cmds:   store.NewCommandQueue(pool),
	}
}

// ID retrievs the backend id
func (b *Backend) ID() string {
	return b.state.ID
}

// Host retrievs the backend host
func (b *Backend) Host() string {
	if b.state.Backend == nil {
		return ""
	}
	return b.state.Backend.Host
}

// Stress calculates the current node load
func (b *Backend) Stress() uint {
	return b.state.MeetingsCount + b.state.AttendeesCount
}

// loadNodeState will fetch all meetings from the backend.
// The meetings are then processed in two passes:
// 1st pass: for each meeting from backend
// If the meeting is in our store there are two cases:
//   - It is assigned to this backend, everything is well.
//     Update meeting stats with info retrieved from the backend.
//   - Otherwise: reassign this meeting to this backend.
// If the meeting is not in our state:
//   - Ignore.
// 2nd pass: for each meeting assigned to backend in store:
// If the meeting is not present in the backend, remove meeting,
// from store
//
// TODO: This function has become way too long.
//       You need to refactor this. Please.
//
func (b *Backend) loadNodeState() error {
	log.Info().
		Str("backend", b.state.Backend.Host).
		Msg("processing backend meetings")

	// Measure latency
	t0 := time.Now()
	req := bbb.GetMeetingsRequest(bbb.Params{}).WithBackend(b.state.Backend)
	rep, err := b.client.Do(req)
	if err != nil {
		// The backend does not even respond, we mark this
		// as an error
		errMsg := fmt.Sprintf("%s", err)
		b.state.NodeState = "error"
		b.state.LastError = &errMsg
		if err := b.state.Save(); err != nil {
			log.Error().Err(err).Msg("save backend state")
		}

		return err
	}
	t1 := time.Now()
	latency := t1.Sub(t0)

	res := rep.(*bbb.GetMeetingsResponse)

	if err != nil {
		errMsg := fmt.Sprintf("%s", err)
		b.state.NodeState = "error"
		b.state.LastError = &errMsg
		if err := b.state.Save(); err != nil {
			log.Error().Err(err).Msg("save backend state")
		}
		return err
	}

	if res.Returncode != "SUCCESS" {
		// Update backend state
		errMsg := fmt.Sprintf("%s: %s", res.MessageKey, res.Message)
		b.state.LastError = &errMsg
		b.state.NodeState = "error"
		if err := b.state.Save(); err != nil {
			log.Error().Err(err).Msg("save backend state")
		}
		return err
	}

	// Update state
	b.state.SyncedAt = time.Now().UTC()
	b.state.LastError = nil
	b.state.Latency = latency
	if b.state.AdminState == "ready" {
		b.state.NodeState = "ready"
	}
	if err := b.state.Save(); err != nil {
		return err
	}

	// Update meetings: 1st pass
	backendMeetings := make([]string, 0, len(res.Meetings))
	for _, meeting := range res.Meetings {
		backendMeetings = append(backendMeetings, meeting.InternalMeetingID)
		mstate, err := store.GetMeetingState(b.pool, store.Q().
			Where("meetings.internal_id = ?", meeting.InternalMeetingID))
		if err != nil {
			// This will happen if something goes wrong with the
			// database, so we can break here.
			return err
		}

		if mstate == nil {
			log.Warn().
				Str("meetingID", meeting.MeetingID).
				Str("internalMeetingID", meeting.InternalMeetingID).
				Str("backend", b.state.Backend.Host).
				Msg("unknown meeting received from backend")

			// Create meeting state, associate with frontend later
			mstate = store.InitMeetingState(b.pool, &store.MeetingState{
				ID:         meeting.MeetingID,
				InternalID: meeting.InternalMeetingID,
				BackendID:  &b.state.ID,
			})
		}

		// Update meeting info and save state
		if err := mstate.Meeting.Update(meeting); err != nil {
			return err
		}
		if err := mstate.Save(); err != nil {
			return err
		}
	}

	// Cleanup meetings: 2nd pass
	count, err := store.DeleteOrphanMeetings(
		b.pool, b.state.ID, backendMeetings)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Warn().
			Int("orphans", int(count)).
			Str("backend", b.state.Backend.Host).
			Msg("removed orphan meetings associated with backend")
	}
	return nil
}

// Meeting State Sync: loadMeetingState will make
// a request to the backend with a get meeting info request.
// This is done for a specific meeting e.g. to sync the
// attendees list - however - most of these things should
// already be handled throught the b3scalenoded on the
// backend instance.
func (b *Backend) refreshMeetingState(
	state *store.MeetingState,
) error {
	req := bbb.GetMeetingInfoRequest(bbb.Params{
		"meetingID": state.ID,
	}).WithBackend(b.state.Backend)
	rep, err := b.client.Do(req)
	if err != nil {
		return err
	}
	res := rep.(*bbb.GetMeetingInfoResponse)
	if res.XMLResponse.Returncode != "SUCCESS" {
		return fmt.Errorf("meeting sync error: %v",
			res.XMLResponse.Message)
	}

	// Update meeting state
	state.Meeting = res.Meeting
	state.SyncedAt = time.Now().UTC()
	return state.Save()
}

// BBB API Implementation

func meetingStateFromRequest(
	pool *pgxpool.Pool,
	req *bbb.Request,
) (*store.MeetingState, error) {
	meetingID, ok := req.Params.MeetingID()
	if !ok {
		return nil, fmt.Errorf("meetingID required")
	}
	// Check if meeting does exist
	meetingState, err := store.GetMeetingState(pool, store.Q().
		Where("id = ?", meetingID))
	return meetingState, err
}

// Create a new Meeting
func (b *Backend) Create(req *bbb.Request) (
	*bbb.CreateResponse, error,
) {
	meetingState, err := meetingStateFromRequest(b.pool, req)
	if err != nil {
		return nil, err
	}
	if meetingState != nil {
		// Check if meeting is runnnig
		res, err := b.IsMeetingRunning(bbb.IsMeetingRunningRequest(
			bbb.Params{
				"meetingID": meetingState.ID,
			}))
		if err != nil {
			return nil, err
		}
		if res.XMLResponse.Returncode == "SUCCESS" {
			// We are good here, just return the current meeting
			// state in a synthetic response.
			res := &bbb.CreateResponse{
				XMLResponse: &bbb.XMLResponse{
					Returncode: "SUCCESS",
				},
				Meeting: meetingState.Meeting,
			}
			res.SetStatus(200)
			return res, nil
		}
	}

	// We don't know about the meeting, or is meeting
	// running did not know about the meeting - anyhow
	// we need to create it.
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	createRes := res.(*bbb.CreateResponse)
	// Update or save meeeting state
	if meetingState == nil {
		_, err = b.state.CreateMeetingState(req.Frontend, createRes.Meeting)
		if err != nil {
			return nil, err
		}
	} else {
		// Update state, associate with backend and frontend
		meetingState.Meeting = createRes.Meeting
		meetingState.SyncedAt = time.Now().UTC()
		if err := meetingState.Save(); err != nil {
			return nil, err
		}
	}

	return createRes, nil
}

// Join a meeting
func (b *Backend) Join(
	req *bbb.Request,
) (*bbb.JoinResponse, error) {
	// Joining a meeting is a process entirely handled by the
	// client. Because of a JSESSIONID which is used? I guess?
	// Or maybe the referrer?
	// Just passing through the location response did not work.
	// For the reverseproxy feature we need to fix this.
	// Even if it means tracking JSESSIONID cookie headers
	// or the referrer and injecting them back through some
	// nxing magic on the BBB side.
	req = req.WithBackend(b.state.Backend)

	// Create custom join response
	res := &bbb.JoinResponse{
		XMLResponse: new(bbb.XMLResponse),
	}

	url := req.URL()
	body := TmplRedirect(url)

	res.SetStatus(http.StatusFound)
	res.SetRaw(body)
	res.SetHeader(http.Header{
		"content-type": []string{"text/html"},
		"location":     []string{req.URL()},
	})

	// Dispatch updating the meeing state
	meetingID, _ := req.Params.MeetingID()
	b.cmds.Queue(UpdateMeetingState(&UpdateMeetingStateRequest{
		ID: meetingID,
	}))

	return res, nil
}

// IsMeetingRunning returns the is meeting running state
func (b *Backend) IsMeetingRunning(
	req *bbb.Request,
) (*bbb.IsMeetingRunningResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	isMeetingRunningRes := res.(*bbb.IsMeetingRunningResponse)
	meetingID, _ := req.Params.MeetingID()
	if isMeetingRunningRes.Returncode == "ERROR" {
		// Delete meeting
		store.DeleteMeetingStateByID(b.pool, meetingID)
	}

	return isMeetingRunningRes, err
}

// End a meeting
func (b *Backend) End(req *bbb.Request) (*bbb.EndResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.EndResponse), err
}

// GetMeetingInfo gets the meeting details
func (b *Backend) GetMeetingInfo(
	req *bbb.Request,
) (*bbb.GetMeetingInfoResponse, error) {
	rep, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	res := rep.(*bbb.GetMeetingInfoResponse)

	// Update our meeting in the store
	if res.XMLResponse.Returncode == "SUCCESS" {
		meetingID, _ := req.Params.MeetingID()
		mstate, err := store.GetMeetingState(b.pool, store.Q().
			Where("id = ?", meetingID))
		if err != nil {
			// We only log the error, as this might fail
			// without impacting the service
			log.Error().
				Err(err).
				Str("backend", b.state.Backend.Host).
				Msg("GetMeetingState")
		} else {
			// Update meeting state
			mstate.Meeting = res.Meeting
			mstate.SyncedAt = time.Now().UTC()
			if err := mstate.Save(); err != nil {
				log.Error().
					Err(err).
					Str("backend", b.state.Backend.Host).
					Msg("Save")
			}
		}
	}

	return res, nil
}

// GetMeetings retrieves a list of meetings
func (b *Backend) GetMeetings(
	req *bbb.Request,
) (*bbb.GetMeetingsResponse, error) {
	// Get all meetings from our store associated
	// with the requesting frontend.
	mstates, err := store.GetMeetingStates(b.pool, store.Q().
		Join("frontends ON frontends.id = meetings.frontend_id").
		Where("meetings.backend_id IS NOT NULL").
		Where("frontend.key = ?", req.Frontend))
	if err != nil {
		return nil, err
	}
	meetings := make([]*bbb.Meeting, 0, len(mstates))
	for _, state := range mstates {
		meetings = append(meetings, state.Meeting)
	}

	// Create response with all meetings
	res := &bbb.GetMeetingsResponse{
		XMLResponse: &bbb.XMLResponse{
			Returncode: "SUCCESS",
		},
		Meetings: meetings,
	}

	return res, nil
}

// GetRecordings retrieves a list of recordings
func (b *Backend) GetRecordings(
	req *bbb.Request,
) (*bbb.GetRecordingsResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.GetRecordingsResponse), nil
}

// PublishRecordings publishes a recording
func (b *Backend) PublishRecordings(
	req *bbb.Request,
) (*bbb.PublishRecordingsResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.PublishRecordingsResponse), nil
}

// DeleteRecordings deletes recordings
func (b *Backend) DeleteRecordings(
	req *bbb.Request,
) (*bbb.DeleteRecordingsResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.DeleteRecordingsResponse), nil
}

// UpdateRecordings updates recordings
func (b *Backend) UpdateRecordings(
	req *bbb.Request,
) (*bbb.UpdateRecordingsResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.UpdateRecordingsResponse), nil
}

// GetDefaultConfigXML retrieves the default config xml
func (b *Backend) GetDefaultConfigXML(
	req *bbb.Request,
) (*bbb.GetDefaultConfigXMLResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.GetDefaultConfigXMLResponse), nil
}

// SetConfigXML sets the? config xml
func (b *Backend) SetConfigXML(
	req *bbb.Request,
) (*bbb.SetConfigXMLResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.SetConfigXMLResponse), nil
}

// GetRecordingTextTracks retrievs all text tracks
func (b *Backend) GetRecordingTextTracks(
	req *bbb.Request,
) (*bbb.GetRecordingTextTracksResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.GetRecordingTextTracksResponse), nil
}

// PutRecordingTextTrack adds a text track
func (b *Backend) PutRecordingTextTrack(
	req *bbb.Request,
) (*bbb.PutRecordingTextTrackResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.PutRecordingTextTrackResponse), nil
}
