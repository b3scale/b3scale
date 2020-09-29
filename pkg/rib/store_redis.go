package rib

import (
	"github.com/go-redis/redis/v8"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// A RedisStore is a implemenation of a Store
// interface using redis.
type RedisStore struct {
	state *cluster.State
	rdb   *redis.ClusterClient
}

// NewRedisStore makes a new store with
// a redis host address and a cluster state.
func NewRedisStore(state *cluster.State, opts *redis.ClusterOptions) *RedisStore {
	rdb := redis.NewClusterClient(opts)
	return &RedisStore{
		state: state,
		rdb:   rdb,
	}
}
