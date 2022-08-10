package api

import (
	"context"
	"net/http"
	"time"

	"github.com/b3scale/b3scale/pkg/store"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// ResourceAgentHeartbeat is the resource for receiving
// an agent heartbeat
var ResourceAgentHeartbeat = &Resource{
	Create: RequireScope(
		ScopeNode,
	)(apiAgentHeartbeat),
}

// AgentHearbeat is a custom API response
type AgentHearbeat struct {
	BackendID string    `json:"backend_id"`
	Heartbeat time.Time `json:"heartbeat"`
}

// Update the backends agent heartbeat
func apiAgentHeartbeat(
	ctx context.Context,
	api *API,
) error {
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		log.Err(err).Msg("could not start transaction")
		return err
	}
	defer tx.Rollback(ctx)

	// Begin Query
	q := store.Q().Where("agent_ref = ?", api.Ref)
	backend, err := store.GetBackendState(ctx, tx, q)
	if err != nil {
		return err
	}
	if backend == nil {
		return echo.ErrNotFound
	}

	heartbeat, err := backend.UpdateAgentHeartbeat(ctx, tx)
	if err != nil {
		return err
	}

	return api.JSON(http.StatusOK, heartbeat)
}
