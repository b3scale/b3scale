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
}

func TestBackendStateinsert(t *testing.T) {
	conn := connectTest(t)
	rnd := uuid.New().String()
	state := InitBackendState(conn, &BackendState{
		Backend: &bbb.Backend{
			Host:   "testhost" + rnd,
			Secret: "testsecret",
		},
		Tags: []string{"2.0.0", "sip", "testing"},
	})

	id, err := state.insert()
	if err != nil {
		t.Error(err)
	}
	t.Log(id)
	t.Log(state)
}

func TestBackendStateSave(t *testing.T) {
	conn := connectTest(t)

	rnd := uuid.New().String()
	state := InitBackendState(conn, &BackendState{
		Backend: &bbb.Backend{
			Host:   "testhost" + rnd,
			Secret: "testsecret",
		},
	})

	err := state.Save()
	if err != nil {
		t.Error(err)
	}

	if state.CreatedAt.IsZero() {
		t.Error("Expected created at to be set.")
	}

	// Update host
	state.Backend.Host = "newhost" + rnd
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
