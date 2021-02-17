package cluster

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// The Controller interfaces with the state of the cluster
// providing methods for retrieving cluster backends and
// frontends.
//
// The controller subscribes to commands.
type Controller struct {
	cmds *store.CommandQueue
	pool *pgxpool.Pool

	lastStartBackground time.Time
	mtx                 sync.Mutex
}

// NewController will initialize the cluster controller
// with a database connection. A BBB client will be created
// which will be used by the backend instances.
func NewController(pool *pgxpool.Pool) *Controller {
	return &Controller{
		cmds: store.NewCommandQueue(pool),
		pool: pool,
	}
}

// Start the controller
func (c *Controller) Start() {
	log.Info().Msg("starting cluster controller")

	// Jitter startup in case multiple instances are spawned at the same time
	time.Sleep(time.Duration(rand.Float64()) * time.Second) // 0 <= jitter < 1.0

	// Periodically start background tasks, even if they
	// are not triggered through requests
	go func() {
		for {
			log.Debug().Msg("running background task periodic trigger")
			c.StartBackground()
			wait := time.Duration(10.0 + 2.0*rand.Float64())
			time.Sleep(wait * time.Second)
		}
	}()

	// Controller Main Loop
	for {
		// Process commands from queue
		if err := c.cmds.Receive(c.handleCommand); err != nil {
			// We will only reach this code when waiting for
			// commands fails. This can happen when the database
			// is down. So, we log the error and wait a bit.
			log.Error().Err(err).Msg("receive next command")
			time.Sleep(1.0 * time.Second)
		}
	}
}

// StartBackground will be run periodically triggered by
// requests and should only add tasks to the command queue
func (c *Controller) StartBackground() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Debounce calls to this function
	if time.Now().Sub(c.lastStartBackground) < 1*time.Second {
		return
	}

	// Add some jitter
	c.lastStartBackground = time.Now()

	// Dispatch loading of the backend state if the
	// last sync was verly long.
	if err := c.requestSyncStale(); err != nil {
		log.Error().Err(err).Msg("requestSyncStale")
	}

	// Dispatch decommissioning of marked backends
	if err := c.requestBackendDecommissions(); err != nil {
		log.Error().Err(err).Msg("requestBackendDecommissions")
	}

	// Check if there are backends where the noded is
	// not present.
	if err := c.warnOfflineBackends(); err != nil {
		log.Error().Err(err).Msg("warnOfflineBackends")
	}
}

// Command callback handler: Decode the operation and
// run the command specific handler. As this is invoked
// by the CommandQueue, these functions are allowed
// to crash and will be recovered.
func (c *Controller) handleCommand(cmd *store.Command) (interface{}, error) {
	// Invoke command handler
	switch cmd.Action {
	case CmdDecommissionBackend:
		log.Debug().Str("cmd", CmdDecommissionBackend).Msg("EXEC")
		return c.handleDecommissionBackend(cmd)
	case CmdUpdateNodeState:
		log.Debug().Str("cmd", CmdUpdateNodeState).Msg("EXEC")
		return c.handleUpdateNodeState(cmd)
	case CmdUpdateMeetingState:
		log.Debug().Str("cmd", CmdUpdateMeetingState).Msg("EXEC")
		return c.handleUpdateMeetingState(cmd)
	case CmdEndAllMeetings:
		log.Debug().Str("cmd", CmdEndAllMeetings).Msg("EXEC")
		return c.handleEndAllMeetings(cmd)
	}

	return nil, ErrUnknownCommand
}

// Command: DecommissionBackend
// Removes a backend state identified by id from the state
func (c *Controller) handleDecommissionBackend(
	cmd *store.Command,
) (interface{}, error) {
	req := &DecommissionBackendRequest{}
	if err := cmd.FetchParams(req); err != nil {
		return nil, err
	}

	// Get backend for decommissioning
	bstate, err := store.GetBackendState(ctx, store.Q().
		Where("id = ?", req.ID))
	if err != nil {
		return nil, err
	}
	if bstate == nil {
		return false, fmt.Errorf("no such backend: %s", req.ID)
	}

	// So. This is how this goes: We check if the backend
	// has active meetings. If this is the case we abort.
	// However, as the admin state indicates a non ready state
	// the router will not longer select this backend
	// for new meetings - so we are good to go here.
	mstates, err := store.GetMeetingStates(c.pool, store.Q().
		Where("meetings.backend_id = ?", req.ID).
		Where("meetings.state->'Running' = ?", true))
	if err != nil {
		return nil, err
	}

	if len(mstates) > 0 {
		// We have running meetings, so we defer this
		log.Warn().
			Int("meetings_running", len(mstates)).
			Msg("decommission backend deferred, backend has meetings running")
		return false, nil
	}

	// Decommission backend by deleting the state
	// and related meetings
	if err := bstate.Delete(); err != nil {
		return false, err
	}

	return true, nil
}

// Command: UpdateNodeState
func (c *Controller) handleUpdateNodeState(
	cmd *store.Command,
) (interface{}, error) {
	// Get backend from command
	req := &UpdateNodeStateRequest{}
	if err := cmd.FetchParams(req); err != nil {
		return nil, err
	}
	backend, err := c.GetBackend(
		store.Q().Where("id = ?", req.ID))
	if err != nil {
		return false, err
	}
	if backend == nil {
		return false, fmt.Errorf("backend not found: %s", req.ID)
	}
	err = backend.loadNodeState()
	if err != nil {
		return false, err
	}
	return true, nil
}

// handleUpdateMeetingState syncs the meeting state
// from a backend
func (c *Controller) handleUpdateMeetingState(
	cmd *store.Command,
) (interface{}, error) {
	req := &UpdateMeetingStateRequest{}
	if err := cmd.FetchParams(req); err != nil {
		return nil, err
	}

	// Get meeting from store
	mstate, err := store.GetMeetingState(c.pool, store.Q().
		Where("id = ?", req.ID))
	if err != nil {
		log.Error().Err(err).Msg("GetMeetingState for handleUpdateMeetingState")
		return nil, err
	}
	if mstate == nil {
		log.Info().
			Str("meetingID", req.ID).
			Msg("meeting is already gone; refesh canceled")
		return false, fmt.Errorf("meeting not found: %s", req.ID)
	}

	// Get Backend
	backend, err := c.GetBackend(
		store.Q().Where("id = ?", mstate.BackendID))
	if err != nil {
		log.Error().Err(err).Msg("GetBackend")
		return nil, err
	}
	if backend == nil {
		log.Info().
			Str("meetingID", req.ID).
			Msg("backend is gone; refesh canceled")
		return false, fmt.Errorf("backend not found")
	}

	// Refresh state
	if err := backend.refreshMeetingState(mstate); err != nil {
		return false, err
	}

	return true, nil
}

// handleEndAllMeetings will send an end request
// for all meetings on a backend
func (c *Controller) handleEndAllMeetings(cmd *store.Command) (interface{}, error) {
	req := &EndAllMeetingsRequest{}
	if err := cmd.FetchParams(req); err != nil {
		return nil, err
	}

	backend, err := c.GetBackend(ctx, store.Q().Where("id = ?", req.BackendID))
	if err != nil {
		return nil, err
	}
	if backend == nil {
		return false, fmt.Errorf("no such backend: %s", req.BackendID)
	}

	// Send end for all *known* meetings. (We however, should *know*
	// all meetings on the backend after a while.)
	mstates, err := store.GetMeetingStates(ctx, store.Q().
		Where("backend_id = ?", req.BackendID))
	if err != nil {
		return nil, err
	}

	for _, m := range mstates {
		log.Info().
			Str("backendID", req.BackendID).
			Str("meetingID", m.Meeting.MeetingID).
			Msg("force end meeting")

		req := bbb.EndRequest(bbb.Params{
			"meetingID": m.Meeting.MeetingID,
			"password":  m.Meeting.ModeratorPW,
		})
		res, err := backend.End(req)
		if err != nil {
			return nil, err
		}

		if res.Returncode != bbb.RetSuccess {
			log.Error().
				Str("meetingID", m.Meeting.MeetingID).
				Str("msg", res.Message).
				Str("msgKey", res.MessageKey).
				Msg("end meeting failed")
			return nil, fmt.Errorf("end meeting failed: %s", res.MessageKey)
		}
	}

	return true, nil
}

// Internal command generators

// requestSyncStale triggers a background sync of the
// entire node state
func (c *Controller) requestSyncStale(ctx context.Context) error {
	log.Debug().Msg("starting stale node refresh")
	stale, err := c.GetBackends(ctx, store.Q().
		Where(`now() - COALESCE(
				synced_at,
				TIMESTAMP '0001-01-01 00:00:00') > ?`,
			time.Duration(10*time.Second)).
		Where("admin_state <> ?", "init"))
	if err != nil {
		return err
	}
	// For each stale backend create a new update state
	// request, which will try to reach the backend.
	for _, b := range stale {
		log.Debug().
			Str("cmd", "UpdateNodeState").
			Str("id", b.state.ID).
			Msg("DISPATCH")
		if err := c.cmds.Queue(
			UpdateNodeState(&UpdateNodeStateRequest{
				ID: b.state.ID,
			})); err != nil {
			return err
		}
	}
	return nil
}

// requestBackendDecommissions will request a decommissioning
// of a backend for all backends, which admin state is marked
// as decommissioned.
func (c *Controller) requestBackendDecommissions(ctx context.Context) error {
	// Get backend states to decommission
	states, err := store.GetBackendStates(ctx, store.Q().
		Where("admin_state = ?", "decommissioned"))
	if err != nil {
		log.Error().Err(err).Msg("decommissioning GetBackendStates")
	}

	if len(states) == 0 {
		return nil // nothing to do here.
	}

	// Decommission backends
	for _, s := range states {
		log.Info().
			Int("count", len(states)).
			Str("host", s.Backend.Host).
			Str("backendID", s.ID).
			Msg("requesting backend decommissioning")

		if err := c.cmds.Queue(
			DecommissionBackend(&DecommissionBackendRequest{
				ID: s.ID,
			})); err != nil {

		}
	}

	return nil
}

// warnOfflineBackends iterates through all unlocked
// backends and warns the user that there are backends offline
func (c *Controller) warnOfflineBackends(ctx context.Context) error {
	// Get offline backends
	deadline := time.Now().UTC().Add(-5 * time.Second)
	states, err := store.GetBackendStates(ctx, store.Q().
		Where("backends.agent_heartbeat < ?", deadline))
	if err != nil {
		return err
	}

	for _, s := range states {
		log.Warn().
			Str("backendID", s.ID).
			Str("host", s.Backend.Host).
			Msg("noded is not available on the backend host")
	}

	return nil
}

// GetBackends retrives backends with a store query
func (c *Controller) GetBackends(
	ctx context.Context,
	q sq.SelectBuilder,
) ([]*Backend, error) {
	states, err := store.GetBackendStates(ctx, q)
	if err != nil {
		return nil, err
	}
	// Make cluster backend from each state
	backends := make([]*Backend, 0, len(states))
	for _, s := range states {
		backends = append(backends, NewBackend(c.pool, s))
	}

	return backends, nil
}

// GetBackend retrievs a single backend by query criteria
func (c *Controller) GetBackend(
	ctx context.Context,
	q sq.SelectBuilder,
) (*Backend, error) {
	backends, err := c.GetBackends(ctx, q)
	if err != nil {
		return nil, err
	}
	if len(backends) == 0 {
		return nil, nil
	}
	return backends[0], nil
}

// GetFrontends retrieves all frontends from
// the store matchig a query
func (c *Controller) GetFrontends(
	ctx context.Context,
	q sq.SelectBuilder,
) ([]*Frontend, error) {
	states, err := store.GetFrontendStates(ctx, q)
	if err != nil {
		return nil, err
	}
	// Make cluster backend from each state
	frontends := make([]*Frontend, 0, len(states))
	for _, s := range states {
		frontends = append(frontends, NewFrontend(s))
	}

	return frontends, nil
}

// GetFrontend fetches a frontend with a state from
// the store
func (c *Controller) GetFrontend(ctx context.Context, q sq.SelectBuilder) (*Frontend, error) {
	frontends, err := c.GetFrontends(ctx, q)
	if err != nil {
		return nil, err
	}
	if len(frontends) == 0 {
		return nil, nil
	}
	return frontends[0], nil
}

// GetMeetingStateByID fetches a meeting state from the store
func (c *Controller) GetMeetingStateByID(ctx context.Context, id string) (*store.MeetingState, error) {
	state, err := store.GetMeetingState(ctx, store.Q().
		Where("id = ?", id))
	if err != nil {
		return nil, err
	}

	return state, nil
}

// DeleteMeetingStateByID purges all knowelege of a meeting
// identified by its ID. If the meeting is unknown, no error
// is raised.
func (c *Controller) DeleteMeetingStateByID(ctx context.Context, id string) error {
	return store.DeleteMeetingStateByInternalID(ctx, id)
}

// BeginTx starts a new transaction in the pool
func (c *Controller) BeginTx(ctx context.Context) pgx.Tx {
	return c.pool.Begin(ctx)
}
