package store

import (
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The FrontendState holds shared information about
// a frontend.
type FrontendState struct {
	ID       string
	Frontend *bbb.Frontend

	CreatedAt time.Time
	UpdatedAt *time.Time
	SyncedAt  *time.Time

	pool *pgxpool.Pool
}
