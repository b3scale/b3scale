package cluster

import (
	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

// A Frontend is a consumer like greenlight.
// Each frontend has it's own secret for authentication.
type Frontend struct {
	ID string

	config *config.Frontend
}

// NewFrontend initializes a frontend with the provided
// config and assigns the ID.
func NewFrontend(config *config.Frontend) *Frontend {
	return &Frontend{
		ID:     config.Key,
		config: config,
	}
}
