package rib

import (
	"github.com/go-redis/redis/v8"
)

// A RedisStore is a implemenation of a Store
// interface using redis.
type RedisStore struct {
	rdb *redis.Client
}
