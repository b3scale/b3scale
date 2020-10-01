package cluster

import (
	"testing"
	// "gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

func makeClusterState() *State {
	state := &State{
		backends: []*Backend{
			&Backend{ID: "backend1"},
			&Backend{ID: "backend2"},
		},
		frontends: []*Frontend{
			&Frontend{ID: "frontend1"},
			&Frontend{ID: "frontend2"},
		},
	}
	return state
}

func TestGetBackend(t *testing.T) {
	state := makeClusterState()

}
