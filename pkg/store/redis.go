package store

/*
 The redis store implements the Store interface
 for holding a shared state across instances.
*/

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// RedisStore is using the redis client for
// storing and retrieving data.
type RedisStore struct {
	rdb *redis.Client
}

// NewRedisStore creates a new instance of a
// redis store.
func NewRedisStore(opts *redis.Options) *RedisStore {
	return &RedisStore{
		rdb: redis.NewClient(opts),
	}
}

// Get retrievs a single value from the store
func (s *RedisStore) Get(key string) (string, error) {
	ctx := context.Background().WithTimeout(30 * time.Second)
	// Lookup meeting with key
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		// Ignore if the key was not found
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	return val, nil
}

// GetAll retrievs all values matching the query
func (s *RedisStore) GetAll(query string) ([]string, error) {
	var (
		results []string
		cursor  uint64
		err     error
	)
	keys := []string{}
	// First get all matching keys
	ctx := context.Background().WithTimeout(30 * time.Second)
	for {
		results, cursor, err = s.rdb.Scan(
			ctx, cursor, query, 100).Result()
		if err != nil {
			return nil, err
		}
		if cursor == 0 {
			break
		}
		keys = append(keys, results...)
	}

	// Retrieve data
	values := make([]string, 0, len(keys))
	for _, k := range keys {
		v, err := s.Get(k)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}

	return values, nil
}

// Store a value in redis
func Set(key, value string) error {
	ctx := context.Background().WithTimeout(30 * time.Second)
	return s.rdb.Set(
		ctx, key, value, 0).Err()
}
