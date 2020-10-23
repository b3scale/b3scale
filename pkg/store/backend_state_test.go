package store

/*
 Backend State Tests
*/

import (
	"testing"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

func TestGetBackendStateByID(t *testing.T) {
	conn := connectTest(t)
	state := InitBackendState(conn, &BackendState{
		Backend: &bbb.Backend{
			Host:   "testhost",
			Secret: "testsecret",
		},
		Tags: []string{"2.0.0", "sip", "testing"},
	})
	err := state.Save()
	if err != nil {
		t.Error("save failed:", err)
	}
}

func TestBackendStateinsert(t *testing.T) {
	conn := connectTest(t)
	state := InitBackendState(conn, &BackendState{
		Backend: &bbb.Backend{
			Host:   "testhost",
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

	state := InitBackendState(conn, &BackendState{
		Backend: &bbb.Backend{
			Host:   "testhost",
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
	state.Backend.Host = "newhost"
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