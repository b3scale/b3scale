package cluster

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

const (
	// MeetingSyncInterval is the amount of time after a
	// meeting state is considered stale and should be refreshed.
	MeetingSyncInterval = 15 * time.Second

	// NodeSyncInterval is the amount of time after a backend
	// node is considered stale and should be refreshed.
	NodeSyncInterval = 20 * time.Second
)

// The Controller interfaces with the state of the cluster
// providing methods for retrieving cluster backends and
// frontends.
//
// The controller subscribes to commands.
type Controller struct {
	cmds *store.CommandQueue

	lastStartBackground time.Time
	mtx                 sync.Mutex
}

// NewController will initialize the cluster controller
// with a database connection. A BBB client will be created
// which will be used by the backend instances.
func NewController() *Controller {
	return &Controller{
		cmds: store.NewCommandQueue(),
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
// requests and should only add tasks to the command queue.
// These tasks will take care of syncing the backends with
// our state by refreshing nodes and meetings.
func (c *Controller) StartBackground() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// Debounce calls to this function
	if time.Now().Sub(c.lastStartBackground) < 10*time.Second {
		return
	}

	c.lastStartBackground = time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := store.Acquire(ctx)
	if err != nil {
		log.Error().Err(err).Msg("could not acquire connection")
		return
	}
	defer conn.Release()
	ctx = store.ContextWithConnection(ctx, conn)

	// Dispatch loading of the backend state if the
	// last sync was verly long.
	if err := c.requestSyncStaleNodes(ctx); err != nil {
		log.Error().Err(err).Msg("requestSyncStale")
	}

	// Dispatch refreshing stale meetings if the last
	// sync was a while ago.
	/*
		CAVEAT:
		This puts a high pressure on the system and might
		not be necessary. For now this is disabled.

		if err := c.requestSyncStaleMeetings(ctx); err != nil {
			log.Error().Err(err).Msg("requestSyncStaleMeetings")
		}
	*/

	// Dispatch decommissioning of marked backends
	if err := c.requestBackendDecommissions(ctx); err != nil {
		log.Error().Err(err).Msg("requestBackendDecommissions")
	}

	// Check if there are backends where the noded is
	// not present.
	if err := c.warnOfflineBackends(ctx); err != nil {
		log.Error().Err(err).Msg("warnOfflineBackends")
	}
}

// Command callback handler: Decode the operation and
// run the command specific handler. As this is invoked
// by the CommandQueue, these functions are allowed
// to crash and will be recovered.
func (c *Controller) handleCommand(
	ctx context.Context,
	cmd *store.Command,
) (interface{}, error) {
	// Invoke command handler
	switch cmd.Action {
	case CmdDecommissionBackend:
		log.Debug().Str("cmd", CmdDecommissionBackend).Msg("EXEC")
		return c.handleDecommissionBackend(ctx, cmd)
	case CmdUpdateNodeState:
		log.Debug().Str("cmd", CmdUpdateNodeState).Msg("EXEC")
		return c.handleUpdateNodeState(ctx, cmd)
	case CmdUpdateMeetingState:
		log.Debug().Str("cmd", CmdUpdateMeetingState).Msg("EXEC")
		return c.handleUpdateMeetingState(ctx, cmd)
	case CmdEndAllMeetings:
		log.Debug().Str("cmd", CmdEndAllMeetings).Msg("EXEC")
		return c.handleEndAllMeetings(ctx, cmd)
	default:
		return nil, ErrUnknownCommand
	}
}

// Command: DecommissionBackend
// Removes a backend state identified by id from the state
func (c *Controller) handleDecommissionBackend(
	ctx context.Context,
	cmd *store.Command,
) (interface{}, error) {
	req := &DecommissionBackendRequest{}
	if err := cmd.FetchParams(ctx, req); err != nil {
		return nil, err
	}

	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get backend for decommissioning
	bstate, err := store.GetBackendState(ctx, tx, store.Q().
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
	mstates, err := store.GetMeetingStates(ctx, tx, store.Q().
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
	if err := bstate.Delete(ctx, tx); err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	return true, nil
}

// Command: UpdateNodeState
func (c *Controller) handleUpdateNodeState(
	ctx context.Context,
	cmd *store.Command,
) (interface{}, error) {
	// Get backend from command
	req := &UpdateNodeStateRequest{}
	if err := cmd.FetchParams(ctx, req); err != nil {
		return nil, err
	}

	backend, err := GetBackend(ctx, store.Q().
		Where("id = ?", req.ID))
	if err != nil {
		return false, err
	}
	if backend == nil {
		return false, fmt.Errorf("backend not found: %s", req.ID)
	}

	err = backend.refreshNodeState(ctx)
	if err != nil {
		return false, err
	}

	return true, nil
}

// handleUpdateMeetingState syncs the meeting state
// from a backend
func (c *Controller) handleUpdateMeetingState(
	ctx context.Context,
	cmd *store.Command,
) (interface{}, error) {
	req := &UpdateMeetingStateRequest{}
	if err := cmd.FetchParams(ctx, req); err != nil {
		return nil, err
	}

	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get meeting from store
	mstate, err := store.GetMeetingState(ctx, tx, store.Q().
		Where("id = ?", req.ID))
	if err != nil {
		log.Error().
			Err(err).
			Msg("GetMeetingState for handleUpdateMeetingState")
		return nil, err
	}
	if mstate == nil {
		log.Debug().
			Str("meetingID", req.ID).
			Msg("meeting is already gone; refesh canceled")
		return false, nil
	}

	// Debounce: Do not refresh the meeting state more than
	// once in 15 seconds (MeetingSyncInterval)
	if !mstate.IsStale(MeetingSyncInterval) {
		return false, nil
	}

	// Get Backend
	backend, err := GetBackend(ctx, store.Q().
		Where("id = ?", mstate.BackendID))
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
	if err := backend.refreshMeetingState(ctx, mstate); err != nil {
		return false, err
	}

	return true, nil
}

// handleEndAllMeetings will send an end request
// for all meetings on a backend
func (c *Controller) handleEndAllMeetings(
	ctx context.Context,
	cmd *store.Command,
) (interface{}, error) {
	req := &EndAllMeetingsRequest{}
	if err := cmd.FetchParams(ctx, req); err != nil {
		return nil, err
	}

	backend, err := GetBackend(ctx, store.Q().
		Where("id = ?", req.BackendID))
	if err != nil {
		return nil, err
	}
	if backend == nil {
		return false, fmt.Errorf("no such backend: %s", req.BackendID)
	}

	// Send end for all *known* meetings. (We however, should *know*
	// all meetings on the backend after a while.)

	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	mstates, err := store.GetMeetingStates(ctx, tx, store.Q().
		Where("backend_id = ?", req.BackendID))
	if err != nil {
		return nil, err
	}
	tx.Rollback(ctx) // We should not block the connection any longer

	for _, m := range mstates {
		log.Info().
			Str("backendID", req.BackendID).
			Str("meetingID", m.Meeting.MeetingID).
			Msg("force end meeting")

		req := bbb.EndRequest(bbb.Params{
			"meetingID": m.Meeting.MeetingID,
			"password":  m.Meeting.ModeratorPW,
		})
		res, err := backend.End(ctx, req)
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

// requestSyncStaleNodes triggers a background sync of the
// entire node state
func (c *Controller) requestSyncStaleNodes(ctx context.Context) error {
	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	log.Debug().Msg("starting stale node refresh")
	stale, err := store.GetBackendStates(ctx, tx, store.Q().
		Where(`now() - COALESCE(
				synced_at,
				TIMESTAMP '0001-01-01 00:00:00') > ?`,
			NodeSyncInterval).
		Where("admin_state <> ?", "init"))
	if err != nil {
		return err
	}
	// For each stale backend create a new update state
	// request, which will try to reach the backend.
	for _, b := range stale {
		log.Debug().
			Str("cmd", "UpdateNodeState").
			Str("id", b.ID).
			Msg("DISPATCH")
		if err := store.QueueCommand(ctx, tx,
			UpdateNodeState(&UpdateNodeStateRequest{
				ID: b.ID,
			})); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// requestSyncStaleMeetings triggers a background sync
// for meetings that have not been synced in a while.
func (c *Controller) requestSyncStaleMeetings(ctx context.Context) error {
	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	log.Debug().Msg("starting stale meeting refresh")
	stale, err := store.GetMeetingStates(ctx, tx, store.Q().
		Where(`now() - COALESCE(
				synced_at,
				TIMESTAMP '0001-01-01 00:00:00') > ?`,
			MeetingSyncInterval))
	if err != nil {
		return err
	}

	// For each stale meeting create a refresh request.
	for _, meeting := range stale {
		log.Debug().
			Str("cmd", "UpdateMeetingState").
			Str("id", meeting.ID).
			Msg("DISPATCH")
		if err := store.QueueCommand(ctx, tx,
			UpdateMeetingState(&UpdateMeetingStateRequest{
				ID: meeting.ID,
			})); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// requestBackendDecommissions will request a decommissioning
// of a backend for all backends, which admin state is marked
// as decommissioned.
func (c *Controller) requestBackendDecommissions(ctx context.Context) error {
	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get backend states to decommission
	states, err := store.GetBackendStates(ctx, tx, store.Q().
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

		if err := store.QueueCommand(ctx, tx,
			DecommissionBackend(&DecommissionBackendRequest{
				ID: s.ID,
			})); err != nil {
			log.Error().
				Err(err).
				Msg("could not queue decommission backend request")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().
			Err(err).
			Msg("could not commit requestBackendDecommissions")
		return err
	}

	return nil
}

// warnOfflineBackends iterates through all unlocked
// backends and warns the user that there are backends offline
func (c *Controller) warnOfflineBackends(ctx context.Context) error {
	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get offline backends
	deadline := time.Now().UTC().Add(-5 * time.Second)
	states, err := store.GetBackendStates(ctx, tx, store.Q().
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
