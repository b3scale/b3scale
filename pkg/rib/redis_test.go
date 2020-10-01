package rib

import (
	"testing"

	// "gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

func makeClusterState() *cluster.State {
	state := cluster.NewState()
	state.AddBackend(&cluster.Backend{ID: "backend1"})
	state.AddBackend(&cluster.Backend{ID: "backend2"})
	state.AddFrontend(&cluster.Frontend{ID: "frontend1"})
	state.AddFrontend(&cluster.Frontend{ID: "frontend2"})
	return state
}

func TestGetBackend(t *testing.T) {
	state := makeClusterState()
}
