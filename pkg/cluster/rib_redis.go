package rib

import (
	//	"context"
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
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

// GetBackend retrieves a backend associated with
// a Meeting from the store.
// If no meeting was found, nil will be returned.
func (s *RedisClusterStore) GetBackend(
	meeting *bbb.Meeting,
) (*cluster.Backend, error) {
	// Check identifier
	if meeting.MeetingID == "" {
		return nil, fmt.Errorf("meeting id is empty")
	}
	ctx := context.Background()
	// Lookup backend id in cache
	id, err := s.rdb.Get(ctx, meeting.MeetingID).Result()
	if err != nil {
		return nil, err
	}
	if id == "" {
		return nil, nil
	}
	backend := s.state.GetBackendByID(id)
	return backend, nil
}

// SetBackend associates a backend with a meeting
func (s *RedisClusterStore) SetBackend(
	meeting *bbb.Meeting, backend *cluster.Backend,
) error {
	// Check identifiers
	if meeting.MeetingID == "" {
		return fmt.Errorf("meeting id is empty")
	}
	if backend.ID == "" {
		return fmt.Errorf("backend id is empty")
	}

	ctx := context.Background()
	return s.rdb.Set(ctx, meeting.MeetingID, backend.ID, 0).Err()
}
