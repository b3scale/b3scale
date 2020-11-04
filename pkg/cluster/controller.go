package cluster

import (
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	// "gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// The Controller interfaces with the state of the cluster
// providing methods for retrieving cluster backends and
// frontends.
//
// The controller subscribes to commands.
type Controller struct {
	cmds *store.CommandQueue
	conn *pgxpool.Pool
}

// NewController will initialize the cluster controller
// with a database connection. A BBB client will be created
// which will be used by the backend instances.
func NewController(conn *pgxpool.Pool) *Controller {
	return &Controller{
		cmds: store.NewCommandQueue(conn),
		conn: conn,
	}
}

// Start the controller
func (c *Controller) Start() {
	log.Println("Starting cluster controller")

	// Create background tasks
	c.StartBackground()

	// Controller Main Loop
	for {
		// Process commands from queue
		if err := c.cmds.Receive(c.handleCommand); err != nil {
			// We will only reach this code when waiting for
			// commands fails. This can happen when the database
			// is down. So, we log the error and wait a bit.
			log.Println(err)
			time.Sleep(1 * time.Second)
		}
	}
}

// StartBackground will be run periodically triggered by
// requests and should only add tasks to the command queue
func (c *Controller) StartBackground() {
	// Dispatch loading of the backend state if the
	// last sync was verly long.
	if err := c.requestSyncStale(); err != nil {
		log.Println(err)
	}
}

// Command callback handler: Decode the operation and
// run the command specific handler. As this is invoked
// by the CommandQueue, these functions are allowed
// to crash and will be recovered.
func (c *Controller) handleCommand(cmd *store.Command) (interface{}, error) {
	// Invoke command handler
	switch cmd.Action {
	case CmdAddBackend:
		return c.handleAddBackend(cmd)
	case CmdRemoveBackend:
		return c.handleRemoveBackend(cmd)
	case CmdUpdateNodeState:
		return c.handleUpdateNodeState(cmd)
	}

	return nil, ErrUnknownCommand
}

// Command: AddBackend
// Creates a new backend state and dispatches the initial
// load state.
func (c *Controller) handleAddBackend(
	cmd *store.Command,
) (interface{}, error) {
	req := &AddBackendRequest{}
	if err := cmd.FetchParams(req); err != nil {
		return nil, err
	}

	// Create new backend state
	state := store.InitBackendState(c.conn, &store.BackendState{
		Backend: req.Backend,
		Tags:    req.Tags,
	})
	if err := state.Save(); err != nil {
		return nil, err
	}

	// Dispatch background job: Load instance state.
	if err := c.cmds.Queue(
		UpdateNodeState(&UpdateNodeStateRequest{
			ID: state.ID,
		})); err != nil {

		return nil, err
	}

	return req.Backend, nil
}

// Command: RemoveBackend
// Removes a backend state identified by id from the state
func (c *Controller) handleRemoveBackend(
	cmd *store.Command,
) (interface{}, error) {
	backendID := cmd.Params.(string)
	return backendID, fmt.Errorf("implement me")
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
		store.NewQuery().Eq("id", req.ID))
	if err != nil {
		return false, err
	}
	if backend == nil {
		return false, fmt.Errorf("backend not found: %s", req.ID)
	}
	err = backend.loadNodeState()
	if err != nil {
		log.Println(backend.state.Backend.Host, err)
		return false, err
	}
	return true, nil
}

// requestSyncStale triggers a background sync of the
// entire node state
func (c *Controller) requestSyncStale() error {
	stale, err := c.GetBackends(store.NewQuery().
		Gt(`now() - COALESCE(
				synced_at,
				TIMESTAMP '0001-01-01 00:00:00')`,
			time.Duration(10*time.Second)).
		Eq("admin_state", "ready"))
	if err != nil {
		return err
	}
	// For each stale backend create a new update state
	// request, which will try to reach the backend.
	for _, b := range stale {
		if err := c.cmds.Queue(
			UpdateNodeState(&UpdateNodeStateRequest{
				ID: b.state.ID,
			})); err != nil {
			return err
		}
	}
	return nil
}

// GetBackends retrives backends with a store query
func (c *Controller) GetBackends(q *store.Query) ([]*Backend, error) {
	states, err := store.GetBackendStates(c.conn, q)
	if err != nil {
		return nil, err
	}
	// Make cluster backend from each state
	backends := make([]*Backend, 0, len(states))
	for _, s := range states {
		backends = append(backends, NewBackend(s))
	}

	return backends, nil
}

// GetBackend retrievs a single backend by query criteria
func (c *Controller) GetBackend(q *store.Query) (*Backend, error) {
	backends, err := c.GetBackends(q)
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
func (c *Controller) GetFrontends(q *store.Query) ([]*Frontend, error) {
	states, err := store.GetFrontendStates(c.conn, q)
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
func (c *Controller) GetFrontend(q *store.Query) (*Frontend, error) {
	frontends, err := c.GetFrontends(q)
	if err != nil {
		return nil, err
	}
	if len(frontends) == 0 {
		return nil, nil
	}
	return frontends[0], nil
}
