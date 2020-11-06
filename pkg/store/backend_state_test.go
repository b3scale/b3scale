package store

/*
 Backend State Tests
*/

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

func backendStateFactory(pool *pgxpool.Pool) *BackendState {
	state := InitBackendState(pool, &BackendState{
		Backend: &bbb.Backend{
			Host:   "testhost-" + uuid.New().String(),
			Secret: "testsecret",
		},
		Tags: []string{"2.0.0", "sip", "testing"},
	})
	return state
}

func TestGetBackendStateByID(t *testing.T) {
	pool := connectTest(t)
	state := backendStateFactory(pool)
	err := state.Save()
	if err != nil {
		t.Error("save failed:", err)
	}

	dbState, err := GetBackendState(pool, Q().
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
	pool := connectTest(t)
	state := backendStateFactory(pool)
	id, err := state.insert()
	if err != nil {
		t.Error(err)
	}
	t.Log(id)
	t.Log(state)
}

func TestBackendStateSave(t *testing.T) {
	pool := connectTest(t)
	state := backendStateFactory(pool)
	err := state.Save()
	if err != nil {
		t.Error(err)
	}

	if state.CreatedAt.IsZero() {
		t.Error("Expected created at to be set.")
	}

	// Update host
	state.Backend.Host = "newhost" + uuid.New().String()
	err = state.Save()
	if err != nil {
		t.Error(err)
	}

	if state.UpdatedAt == nil {
		t.Error("Update date should bet set.")
	}
	t.Log(state.UpdatedAt)
	t.Log(state)
}

func TestCreateMeeting(t *testing.T) {
	pool := connectTest(t)
	bstate := backendStateFactory(pool)
	if err := bstate.Save(); err != nil {
		t.Error(err)
		return
	}
	fstate := frontendStateFactory(pool)
	if err := fstate.Save(); err != nil {
		t.Error(err)
		return
	}

	// Create meeting state
	mstate, err := bstate.CreateMeetingState(fstate.Frontend, &bbb.Meeting{
		MeetingID:   uuid.New().String(),
		MeetingName: "foo",
	})
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(mstate.ID)
}
