package v1

import (
	"context"
	"strings"

	"github.com/b3scale/b3scale/pkg/store"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
)

// BackendFromQuery resolves the backend, either identified by ID or by
// hostname. The hostname must be an exact match.
func BackendFromQuery(
	ctx context.Context,
	api *APIContext,
	tx pgx.Tx,
) (*store.BackendState, error) {
	id := strings.TrimSpace(api.QueryParam("backend_id"))
	host := strings.TrimSpace(api.QueryParam("backend_host"))

	hasQuery := false
	q := store.Q()
	if id != "" {
		q = q.Where("id = ?", id)
		hasQuery = true
	}
	if host != "" {
		q = q.Where("host = ?", host)
		hasQuery = true
	}
	if !hasQuery {
		return nil, echo.ErrBadRequest
	}

	backend, err := store.GetBackendState(ctx, tx, q)
	if err != nil {
		return nil, err
	}
	return backend, nil
}
