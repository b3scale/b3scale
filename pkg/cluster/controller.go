package cluster

import (
	"log"

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
	conn  *pgxpool.Pool
	state *store.ClusterState
}

// NewController will initialize the cluster controller
// with a database connection. A BBB client will be created
// which will be used by the backend instances.
func NewController(state *store.ClusterState, conn *pgxpool.Pool) *Controller {
	return &Controller{
		state: state,
		conn:  conn,
	}
}

// Start the controller
func (c *Controller) Start() {
	log.Println("Starting cluster controller")

	// Enter command loop
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
