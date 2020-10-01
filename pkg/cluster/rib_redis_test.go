package cluster

import (
	"testing"

	"github.com/go-redis/redis/v8"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
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

func makeRedisRIB(s *State) *RedisClusterRIB {
	return NewRedisClusterRIB(s, &redis.ClusterOptions{
		Addrs: []string{":6379"},
	})
}

func TestGetBackend(t *testing.T) {
	state := makeClusterState()
	rib := makeRedisRIB(state)

	// Associate meeting
	b1 := state.GetBackendByID("backend1")
	m := &bbb.Meeting{
		MeetingID: "meeting1",
	}

	res, err := rib.GetBackend(m)
	if err != nil {
		t.Error(err)
	}
	if res != nil {
		t.Error("Unexpected result:", res)
	}

	err = rib.SetBackend(m, b1)
	if err != nil {
		t.Error(err)
	}

}
