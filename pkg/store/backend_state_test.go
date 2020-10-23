package store

/*
 Backend State Tests
*/

import (
	"testing"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

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
	t.Log(state)
}
