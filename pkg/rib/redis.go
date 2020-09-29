package rib

import (
	"github.com/go-redis/redis/v8"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// A RedisClusterStore is a implemenation of a Store
// interface using a redis cluster client.
type RedisClusterStore struct {
	state *cluster.State
	rdb   *redis.ClusterClient
}

// NewRedisClusterStore makes a new store with
// a redis host address and a cluster state.
func NewRedisClusterStore(
	state *cluster.State,
	opts *redis.ClusterOptions,
) *RedisClusterStore {
	rdb := redis.NewClusterClient(opts)
	return &RedisClusterStore{
		state: state,
		rdb:   rdb,
	}
}
