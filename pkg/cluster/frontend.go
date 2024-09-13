package cluster

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/store"
)

// A Frontend is a consumer like Greenlight.
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

// ID retrievs the frontend id
func (f *Frontend) ID() string {
	return f.state.ID
}

// Key retrieves the frontend key
func (f *Frontend) Key() string {
	if f.state.Frontend == nil {
		return ""
	}
	return f.state.Frontend.Key
}

// Frontend gets the states BBB frontend
func (f *Frontend) Frontend() *bbb.Frontend {
	return f.state.Frontend
}

// Active gets the disable/enable state for the frontend
func (f *Frontend) Active() bool {
	return f.state.Active
}

// Settings gets the state settings
func (f *Frontend) Settings() *store.FrontendSettings {
	return &f.state.Settings
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

// GetFrontends retrieves all frontends from
// the store matching a query
func GetFrontends(
	ctx context.Context,
	q sq.SelectBuilder,
) ([]*Frontend, error) {
	conn := store.ConnectionFromContext(ctx)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	states, err := store.GetFrontendStates(ctx, tx, q)
	if err != nil {
		return nil, err
	}
	tx.Rollback(ctx)

	// Make cluster backend from each state
	frontends := make([]*Frontend, 0, len(states))
	for _, s := range states {
		frontends = append(frontends, NewFrontend(s))
	}
	return frontends, nil
}

// GetFrontend fetches a frontend with a state from
// the store
func GetFrontend(
	ctx context.Context,
	q sq.SelectBuilder,
) (*Frontend, error) {
	frontends, err := GetFrontends(ctx, q)
	if err != nil {
		return nil, err
	}
	if len(frontends) == 0 {
		return nil, nil
	}
	return frontends[0], nil
}
