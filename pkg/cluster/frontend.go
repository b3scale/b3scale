package cluster

import (
	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

// A Frontend is a consumer like greenlight.
// Each frontend has it's own secret for authentication.
type Frontend struct {
	config *config.Frontend
}
