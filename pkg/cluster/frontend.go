package cluster

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"

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

// GetFrontends retrieves all frontends from
// the store matchig a query
func GetFrontends(
	ctx context.Context,
	tx pgx.Tx,
	q sq.SelectBuilder,
) ([]*Frontend, error) {
	states, err := store.GetFrontendStates(ctx, tx, q)
	if err != nil {
		return nil, err
	}
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
	tx pgx.Tx,
	q sq.SelectBuilder,
) (*Frontend, error) {
	frontends, err := GetFrontends(ctx, tx, q)
	if err != nil {
		return nil, err
	}
	if len(frontends) == 0 {
		return nil, nil
	}
	return frontends[0], nil
}
