package store

/*
 Backend State Tests
*/

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

func backendStateFactory() *BackendState {
	state := InitBackendState(&BackendState{
		Backend: &bbb.Backend{
			Host:   "testhost-" + uuid.New().String(),
			Secret: "testsecret",
		},
		Settings: BackendSettings{
			Tags: []string{"2.0.0", "sip", "testing"},
		},
	})
	return state
}

func TestGetBackendStateByID(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	state := backendStateFactory()
	err := state.Save(context.Background(), tx)
	if err != nil {
		t.Error("save failed:", err)
	}

	dbState, err := GetBackendState(ctx, tx, Q().
		Where("id = ?", state.ID))
	if err != nil {
		t.Error(err)
		return
	}
	if dbState == nil {
		t.Error("did not find backend by id")
	}
	if dbState.ID != state.ID {
		t.Error("unexpected id:", dbState.ID)
	}
}

func TestBackendStateinsert(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	state := backendStateFactory()
	id, err := state.insert(ctx, tx)
	if err != nil {
		t.Error(err)
	}
	t.Log(id)
	t.Log(state)
}

func TestBackendStateSave(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	state := backendStateFactory()
	err := state.Save(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if state.CreatedAt.IsZero() {
		t.Error("Expected created at to be set.")
	}

	// Update host
	state.Backend.Host = "newhost" + uuid.New().String()
	err = state.Save(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	t.Log(state.SyncedAt)
	t.Log(state)
}

func TestCreateMeeting(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	bstate := backendStateFactory()
	if err := bstate.Save(ctx, tx); err != nil {
		t.Error(err)
		return
	}
	fstate := frontendStateFactory()
	if err := fstate.Save(ctx, tx); err != nil {
		t.Error(err)
		return
	}

	// Create meeting state
	mstate, err := bstate.CreateMeetingState(ctx, tx, fstate.Frontend, &bbb.Meeting{
		MeetingID:         uuid.New().String(),
		InternalMeetingID: uuid.New().String(),
		MeetingName:       "foo",
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(mstate.ID)
}

func TestBackendStateAgentHeartbeat(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	state := backendStateFactory()
	err := state.Save(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	// Fresh backen, agent should not be alive
	if state.IsAgentAlive() {
		t.Error("there should never have been a heartbeat")
	}

	// Make heartbeat
	if err := state.UpdateAgentHeartbeat(ctx, tx); err != nil {
		t.Error(err)
	}

	if !state.IsAgentAlive() {
		t.Error("agent should be alive")
	}
}
