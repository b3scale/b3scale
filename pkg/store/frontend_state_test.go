package store

import (
	"testing"

	"github.com/google/uuid"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

func TestFrontendStateSave(t *testing.T) {
	pool := connectTest(t)
	state := InitFrontendState(pool, &FrontendState{
		Frontend: &bbb.Frontend{
			Key:    uuid.New().String(),
			Secret: "v3rys3cr37",
		},
		Active: true,
	})

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
	if state.UpdatedAt != nil {
		t.Error("Unexpected updated at:", state.UpdatedAt)
	}
	state.Active = false
	if err := state.Save(); err != nil {
		t.Error(err)
	}

	if state.UpdatedAt == nil {
		t.Error("Unexpected updated at:", state.UpdatedAt)
	}
}
