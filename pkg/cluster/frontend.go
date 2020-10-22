package cluster

// A Frontend is a consumer like greenlight.
// Each frontend has it's own secret for authentication.
type Frontend struct {
	state *store.FrontendState
}

// NewFrontend initializes a frontend with the provided
// config and assigns the ID.
func NewFrontend(state *store.FrontendState) *Frontend {
	return &Frontend{
		state: state,
	}
}
