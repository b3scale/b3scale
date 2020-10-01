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

func makeRedisRIB(s *State) *RedisRIB {
	return NewRedisRIB(s, &redis.Options{
		Addr: ":6379",
	})
}

// Test redis rib
func TestRedisRIBBackend(t *testing.T) {
	s := makeClusterState()
	rib := makeRedisRIB(s)

	be := &Backend{ID: "backend1"}
	m := &bbb.Meeting{MeetingID: "meeeeeeeeting00000fff"}

	// Clear meeting
	err := rib.Delete(m)
	if err != nil {
		t.Error(err)
	}
	retb, err := rib.GetBackend(m)
	if err != nil {
		t.Error(err)
	}
	if retb != nil {
		t.Error("There should be no associated backend.")
	}
	err = rib.SetBackend(m, be)
	if err != nil {
		t.Error(err)
	}
	retb, err = rib.GetBackend(m)
	if err != nil {
		t.Error(err)
	}
	if retb == nil {
		t.Error("There should be an associated backend!")
	}
	// Delete backend
	err = rib.SetBackend(m, nil)
	if err != nil {
		t.Error(err)
	}
	retb, err = rib.GetBackend(m)
	if err != nil {
		t.Error(err)
	}
	if retb != nil {
		t.Error("There should be no associated backend.")
	}
}

func makeRedisClusterRIB(s *State) *RedisClusterRIB {
	return NewRedisClusterRIB(s, &redis.ClusterOptions{
		Addrs: []string{":6379"},
	})
}

// DISABLED: For now we use a normal redis client
func _TestRedisClusterGetBackend(t *testing.T) {
	state := makeClusterState()
	rib := makeRedisClusterRIB(state)

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
