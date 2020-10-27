package store

import (
	"time"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The FrontendState holds shared information about
// a frontend.
type FrontendState struct {
	ID       string
	Frontend *bbb.Frontend
}
