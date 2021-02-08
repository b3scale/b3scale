package cluster

import (
	"fmt"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

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

// Frontend gets the states BBB frontend
func (f *Frontend) Frontend() *bbb.Frontend {
	return f.state.Frontend
}

// String stringifies the frontend
func (f *Frontend) String() string {
	if f.state != nil {
		key := ""
		if f.state.Frontend != nil {
			key = f.state.Frontend.Key
		}
		return fmt.Sprintf("<Frontend id='%v', key='%v'>", f.state.ID, key)
	}
	return "<Frontend>"
}
