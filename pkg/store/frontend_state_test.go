package store

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

func frontendStateFactory(pool *pgxpool.Pool) *FrontendState {
	state := InitFrontendState(pool, &FrontendState{
		Frontend: &bbb.Frontend{
			Key:    uuid.New().String(),
			Secret: "v3rys3cr37",
		},
		Active: true,
	})
	return state
}

func TestFrontendStateSave(t *testing.T) {
	pool := connectTest(t)
	state := frontendStateFactory(pool)

	// Create / Insert
	if state.ID != "" {
		t.Log("Unexcted empty ID for new state:", state.ID)
	}
	if err := state.Save(); err != nil {
		t.Error()
	}
	if state.ID == "" {
		t.Log("Expected state ID to be assigned.")
	}
	t.Log(state.ID)

	// Update
	state.Active = false
	if err := state.Save(); err != nil {
		t.Error(err)
	}

	if state.UpdatedAt.IsZero() {
		t.Error("Unexpected updated at:", state.UpdatedAt)
	}
}

func TestGetFrontendState(t *testing.T) {
	pool := connectTest(t)
	state := frontendStateFactory(pool)
	if err := state.Save(); err != nil {
		t.Error(err)
	}
	ret, err := GetFrontendState(pool, Q().
		Where("key = ?", state.Frontend.Key))
	if err != nil {
		t.Error(err)
	}
	t.Log(ret)
}
