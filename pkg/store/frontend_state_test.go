package store

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

func frontendStateFactory() *FrontendState {
	state := InitFrontendState(&FrontendState{
		Frontend: &bbb.Frontend{
			Key:    uuid.New().String(),
			Secret: "v3rys3cr37",
		},
		Active: true,
	})
	return state
}

func TestFrontendStateSave(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	state := frontendStateFactory()

	// Create / Insert
	if state.ID != "" {
		t.Log("Unexcted empty ID for new state:", state.ID)
	}
	if err := state.Save(ctx, tx); err != nil {
		t.Error()
	}
	if state.ID == "" {
		t.Log("Expected state ID to be assigned.")
	}
	t.Log(state.ID)

	// Update
	state.Active = false
	if err := state.Save(ctx, tx); err != nil {
		t.Error(err)
	}

	if state.UpdatedAt.IsZero() {
		t.Error("Unexpected updated at:", state.UpdatedAt)
	}
}

func TestGetFrontendState(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	state := frontendStateFactory()
	if err := state.Save(ctx, tx); err != nil {
		t.Error(err)
	}
	ret, err := GetFrontendState(ctx, tx, Q().
		Where("key = ?", state.Frontend.Key))
	if err != nil {
		t.Error(err)
	}
	t.Log(ret)
}

func TestFrontendValidate(t *testing.T) {
	state := frontendStateFactory()
	if err := state.Validate(); err != nil {
		t.Error("frontend state should be valid")
	}

	state = &FrontendState{
		Frontend: &bbb.Frontend{},
	}
	err := state.Validate()
	if err == nil {
		t.Error("validation should have failed")
	}
	t.Log(err)
}
