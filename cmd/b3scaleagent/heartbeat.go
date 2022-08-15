package main

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/http/api/client"
)

// StartHeartbeat will periodically inform b3scale
// about our existance.
func StartHeartbeat(
	ctx context.Context,
	api *client.Client,
) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}
		if _, err := api.AgentHeartbeatCreate(ctx); err != nil {
			log.Error().Err(err).
				Msg("could not create heartbeat")
		}
		time.Sleep(1 * time.Second)
	}
}
