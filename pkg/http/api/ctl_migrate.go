package api

import (
	"context"
	"net/http"

	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/store/schema"
)

// ResourceCtlMigrate is a restful group for applying
// migrations
var ResourceCtlMigrate = &Resource{
	Create: RequireScope(
		auth.ScopeAdmin,
	)(apiCtlMigrate),
}

// Apply all pending migrations
func apiCtlMigrate(
	ctx context.Context,
	api *API,
) error {
	dbURL := config.EnvOpt(config.EnvDbURL, config.EnvDbURLDefault)
	m := schema.NewManager(dbURL)
	if err := m.Migrate(ctx, m.DB); err != nil {
		return err
	}
	status := m.Status(ctx)
	return api.JSON(http.StatusOK, status)
}
