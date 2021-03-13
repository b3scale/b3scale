package cluster

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
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
}

// NewBackend creates a new backend instance with
// a fresh bbb client.
func NewBackend(state *store.BackendState) *Backend {
	return &Backend{
		client: bbb.NewClient(),
		state:  state,
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

// GetBackends retrievs all backends from the store,
// filterable with a query.
func GetBackends(
	ctx context.Context,
	q sq.SelectBuilder,
) ([]*Backend, error) {
	tx, err := store.Begin(ctx)
	if err != nil {
		return nil, err
	}
	states, err := store.GetBackendStates(ctx, tx, q)
	if err != nil {
		return nil, err
	}
	tx.Rollback(ctx)

	// Make cluster backend from each state
	backends := make([]*Backend, 0, len(states))
	for _, s := range states {
		backends = append(backends, NewBackend(s))
	}

	return backends, nil
}

// GetBackend retrievs a single backend by query criteria
func GetBackend(
	ctx context.Context,
	q sq.SelectBuilder,
) (*Backend, error) {
	backends, err := GetBackends(ctx, q)
	if err != nil {
		return nil, err
	}
	if len(backends) == 0 {
		return nil, nil
	}
	return backends[0], nil
}

// Stress calculates the current node load
func (b *Backend) Stress() float64 {
	f := b.state.LoadFactor

	// Assume a base load of n attendees. This should come from
	// some a config or should be set per frontend.
	attendeeBaseLoad := 15.0
	attendeeLoad := math.Max(attendeeBaseLoad, float64(b.state.AttendeesCount))
	return f * (float64(b.state.MeetingsCount) + attendeeLoad)
}

// refreshNodeState will fetch all meetings from the backend.
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
func (b *Backend) refreshNodeState(
	ctx context.Context,
) error {
	log.Debug().
		Str("backend", b.state.Backend.Host).
		Msg("processing backend meetings")

	// Measure latency
	t0 := time.Now()
	req := bbb.GetMeetingsRequest(bbb.Params{}).WithBackend(b.state.Backend)
	rep, err := b.client.Do(ctx, req)
	if err != nil {
		// The backend does not even respond, we mark this
		// as an error, however our refresh was "successful".
		errMsg := fmt.Sprintf("%s", err)
		b.state.NodeState = "error"
		b.state.LastError = &errMsg

		if err := store.BeginFunc(ctx, func(tx pgx.Tx) error {
			return b.state.Save(ctx, tx)
		}); err != nil {
			log.Error().Err(err).Msg("save backend state")
		}
	}
	t1 := time.Now()
	latency := t1.Sub(t0)

	res := rep.(*bbb.GetMeetingsResponse)
	if res.Returncode != "SUCCESS" {
		// Update backend state
		errMsg := fmt.Sprintf("%s: %s", res.MessageKey, res.Message)
		b.state.LastError = &errMsg
		b.state.NodeState = "error"

		if err := store.BeginFunc(ctx, func(tx pgx.Tx) error {
			return b.state.Save(ctx, tx)
		}); err != nil {
			log.Error().Err(err).Msg("save backend state")
		}
	}

	// Update state
	b.state.SyncedAt = time.Now().UTC()
	b.state.LastError = nil
	b.state.Latency = latency
	if b.state.AdminState == "ready" {
		b.state.NodeState = "ready"
	}

	if err := store.BeginFunc(ctx, func(tx pgx.Tx) error {
		return b.state.Save(ctx, tx)
	}); err != nil {
		log.Error().Err(err).Msg("save backend state")
	}

	// Update meetings: 1st pass
	backendMeetings := make([]string, 0, len(res.Meetings))
	for _, meeting := range res.Meetings {
		if err := store.BeginFunc(ctx, func(tx pgx.Tx) error {
			return b.state.CreateOrUpdateMeetingState(ctx, tx, meeting)
		}); err != nil {
			log.Error().
				Err(err).
				Str("meetingID", meeting.MeetingID).
				Str("internalMeetingID", meeting.InternalMeetingID).
				Msg("could not save meeting state during node refresh")
			continue
		}
		backendMeetings = append(backendMeetings, meeting.InternalMeetingID)
	}

	// Cleanup meetings: 2nd pass
	if err := store.BeginFunc(ctx, func(tx pgx.Tx) error {
		count, err := store.DeleteOrphanMeetings(ctx, tx, b.state.ID, backendMeetings)
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
	}); err != nil {
		log.Error().Err(err).Msg("commit delete orphans")
	}

	// Update meetings and attendees counter
	return store.BeginFunc(ctx, func(tx pgx.Tx) error {
		return b.state.UpdateStatCounters(ctx, tx)
	})
}

// Meeting State Sync: loadMeetingState will make
// a request to the backend with a get meeting info request.
// This is done for a specific meeting e.g. to sync the
// attendees list - however - most of these things should
// already be handled throught the b3scalenoded on the
// backend instance.
func (b *Backend) refreshMeetingState(
	ctx context.Context,
	state *store.MeetingState,
) error {
	req := bbb.GetMeetingInfoRequest(bbb.Params{
		"meetingID": state.ID,
	}).WithBackend(b.state.Backend)
	rep, err := b.client.Do(ctx, req)
	if err != nil {
		return err
	}
	res := rep.(*bbb.GetMeetingInfoResponse)
	if res.XMLResponse.Returncode != "SUCCESS" {
		// The meeting could not be found
		// For now, let's remove the meeting from our state
		if err := store.BeginFunc(ctx, func(tx pgx.Tx) error {
			return store.DeleteMeetingStateByInternalID(
				ctx,
				tx,
				state.InternalID)
		}); err != nil {
			return err
		}
	}

	// Update meeting state
	state.Meeting = res.Meeting
	state.MarkSynced()

	return store.BeginFunc(ctx, func(tx pgx.Tx) error {
		return state.Save(ctx, tx)
	})
}

// BBB API Implementation

func meetingStateFromRequest(
	ctx context.Context,
	tx pgx.Tx,
	req *bbb.Request,
) (*store.MeetingState, error) {
	meetingID, ok := req.Params.MeetingID()
	if !ok {
		return nil, fmt.Errorf("meetingID required")
	}
	// Check if meeting does exist
	meetingState, err := store.GetMeetingState(ctx, tx, store.Q().
		Where("id = ?", meetingID))
	return meetingState, err
}

// Version responds with the current version. This request
// will not hit a real backend and is not part of the
// API interface.
func (b *Backend) Version(req *bbb.Request) (*bbb.XMLResponse, error) {
	res := &bbb.XMLResponse{
		Returncode: "SUCCESS",
		Version:    "2.0",
	}
	res.SetStatus(200)
	return res, nil
}

// Create a new Meeting
func (b *Backend) Create(
	ctx context.Context,
	req *bbb.Request,
) (*bbb.CreateResponse, error) {
	// TODO: Do we need to switch context here?
	txctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Update or save meeeting state transaction. We acquire this before
	// doing the backend request so we can make sure it will be commited
	// and safed.
	tx, err := store.Begin(txctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(txctx)

	res, err := b.client.Do(ctx, req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	createRes := res.(*bbb.CreateResponse)
	if createRes.Meeting == nil {
		log.Error().
			Msg("create returned without a meeting")
		return nil, fmt.Errorf("meeting was not created on server")
	}

	meetingState, err := meetingStateFromRequest(txctx, tx, req)
	if err != nil {
		return nil, err
	}
	if meetingState == nil {
		_, err = b.state.CreateMeetingState(txctx, tx, req.Frontend, createRes.Meeting)
		if err != nil {
			return nil, err
		}
	} else {
		// Update state, associate with backend and frontend
		meetingState.Meeting = createRes.Meeting
		meetingState.SyncedAt = time.Now().UTC()
		if err := meetingState.Save(txctx, tx); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(txctx); err != nil {
		return nil, err
	}
	return createRes, nil
}

// Join via redirect: The client will receive a
// redirect to the BBB backend and will join there directly.
func (b *Backend) Join(
	ctx context.Context,
	req *bbb.Request,
) (*bbb.JoinResponse, error) {
	// Joining a meeting is a process entirely handled by the
	// client. Because of a JSESSIONID which is used for preventing
	// session stealing, just passing through the location response
	// does not work. The JSESSIONID cookie is not associtated with
	// the backend domain and thus the sessionToken is not accepted
	// as valid.
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

	return res, nil
}

// JoinProxy makes a request on behalf of the client.
// A reverse proxy needs to pass all subsequent requests
// to the BBB backend.
func (b *Backend) JoinProxy(
	ctx context.Context,
	req *bbb.Request,
) (*bbb.JoinResponse, error) {
	res, err := b.client.Do(ctx, req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	joinRes := res.(*bbb.JoinResponse)
	if joinRes.Status() != 302 {
		return joinRes, nil // Not the expected redirect
	}

	hostURL, _ := url.Parse(b.state.Backend.Host)

	// Rewrite redirect to us, also keep the jsession cookie
	// and add the backend host to the query for pinning
	joinURL, err := url.Parse(joinRes.Header().Get("Location"))
	if err != nil {
		return nil, err
	}
	joinURL.Scheme = ""
	joinURL.Host = ""

	q := joinURL.Query()
	q.Add("b3shost", hostURL.Host)
	joinURL.RawQuery = q.Encode()

	joinRes.Header().Set("Location", joinURL.String())

	return joinRes, nil
}

// IsMeetingRunning returns the is meeting running state
func (b *Backend) IsMeetingRunning(
	ctx context.Context,
	req *bbb.Request,
) (*bbb.IsMeetingRunningResponse, error) {
	res, err := b.client.Do(ctx, req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}

	isMeetingRunningRes := res.(*bbb.IsMeetingRunningResponse)
	meetingID, _ := req.Params.MeetingID()
	if isMeetingRunningRes.Returncode == "ERROR" {
		// Delete meeting
		if err := store.BeginFunc(ctx, func(tx pgx.Tx) error {
			return store.DeleteMeetingStateByID(ctx, tx, meetingID)
		}); err != nil {
			log.Error().
				Err(err).
				Msg("failed to commit delete meeting")
		}
	}
	return isMeetingRunningRes, err
}

// End a meeting
func (b *Backend) End(
	ctx context.Context,
	req *bbb.Request,
) (*bbb.EndResponse, error) {
	res, err := b.client.Do(ctx, req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.EndResponse), err
}

// GetMeetingInfo gets the meeting details
func (b *Backend) GetMeetingInfo(
	ctx context.Context,
	req *bbb.Request,
) (*bbb.GetMeetingInfoResponse, error) {
	rep, err := b.client.Do(ctx, req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	res := rep.(*bbb.GetMeetingInfoResponse)

	// Update our meeting in the store
	if res.XMLResponse.Returncode == "SUCCESS" {
		tx, err := store.Begin(ctx)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)

		meetingID, _ := req.Params.MeetingID()
		mstate, err := store.GetMeetingState(ctx, tx, store.Q().
			Where("id = ?", meetingID))
		if err != nil {
			// We only log the error, as this might fail
			// without impacting the service
			log.Error().
				Err(err).
				Str("backend", b.state.Backend.Host).
				Msg("GetMeetingState")
		} else {
			if mstate == nil {
				log.Warn().
					Str("backend", b.state.Backend.Host).
					Str("meetingID", meetingID).
					Msg("GetMeetingInfo for unknown meeting")
			} else {
				// Update meeting state
				mstate.Meeting = res.Meeting
				mstate.SyncedAt = time.Now().UTC()
				if err := mstate.Save(ctx, tx); err != nil {
					log.Error().
						Err(err).
						Str("backend", b.state.Backend.Host).
						Msg("Save")
				}
			}
		}

		// Persist changes
		if err := tx.Commit(ctx); err != nil {
			log.Error().
				Err(err).
				Msg("failed to commit meeting info")
		}
	}

	return res, nil
}

// GetMeetings retrieves a list of meetings
func (b *Backend) GetMeetings(
	ctx context.Context,
	req *bbb.Request,
) (*bbb.GetMeetingsResponse, error) {
	tx, err := store.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get all meetings from our store associated
	// with the requesting frontend.
	mstates, err := store.GetMeetingStates(ctx, tx, store.Q().
		Join("frontends ON frontends.id = meetings.frontend_id").
		Where("meetings.backend_id IS NOT NULL").
		Where("frontends.key = ?", req.Frontend.Key))
	if err != nil {
		return nil, err
	}

	tx.Rollback(ctx)

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
	ctx context.Context,
	req *bbb.Request,
) (*bbb.GetRecordingsResponse, error) {
	res, err := b.client.Do(ctx, req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.GetRecordingsResponse), nil
}

// PublishRecordings publishes a recording
func (b *Backend) PublishRecordings(
	ctx context.Context,
	req *bbb.Request,
) (*bbb.PublishRecordingsResponse, error) {
	res, err := b.client.Do(ctx, req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.PublishRecordingsResponse), nil
}

// DeleteRecordings deletes recordings
func (b *Backend) DeleteRecordings(
	ctx context.Context,
	req *bbb.Request,
) (*bbb.DeleteRecordingsResponse, error) {
	res, err := b.client.Do(ctx, req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.DeleteRecordingsResponse), nil
}

// UpdateRecordings updates recordings
func (b *Backend) UpdateRecordings(
	ctx context.Context,
	req *bbb.Request,
) (*bbb.UpdateRecordingsResponse, error) {
	res, err := b.client.Do(ctx, req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.UpdateRecordingsResponse), nil
}

// GetDefaultConfigXML retrieves the default config xml
func (b *Backend) GetDefaultConfigXML(
	ctx context.Context,
	req *bbb.Request,
) (*bbb.GetDefaultConfigXMLResponse, error) {
	res, err := b.client.Do(ctx, req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.GetDefaultConfigXMLResponse), nil
}

// SetConfigXML sets the? config xml
func (b *Backend) SetConfigXML(
	ctx context.Context,
	req *bbb.Request,
) (*bbb.SetConfigXMLResponse, error) {
	res, err := b.client.Do(ctx, req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.SetConfigXMLResponse), nil
}

// GetRecordingTextTracks retrievs all text tracks
func (b *Backend) GetRecordingTextTracks(
	ctx context.Context,
	req *bbb.Request,
) (*bbb.GetRecordingTextTracksResponse, error) {
	res, err := b.client.Do(ctx, req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.GetRecordingTextTracksResponse), nil
}

// PutRecordingTextTrack adds a text track
func (b *Backend) PutRecordingTextTrack(
	ctx context.Context,
	req *bbb.Request,
) (*bbb.PutRecordingTextTrackResponse, error) {
	res, err := b.client.Do(ctx, req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.PutRecordingTextTrackResponse), nil
}

// String stringifies the Backend
func (b *Backend) String() string {
	if b.state != nil {
		host := "unknown host"
		if b.state.Backend != nil {
			host = b.state.Backend.Host
		}
		return fmt.Sprintf("<Backend id='%v' host='%v'>", b.state.ID, host)
	}
	return "<Backend>"
}
