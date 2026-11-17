package sessionstore

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/os/gsession"
)

// RedisStorage uses GoFrame's Redis hash session storage and honors the TTL
// supplied by the server session configuration.
type RedisStorage struct {
	*gsession.StorageRedisHashTable
}

func NewRedisStorage(redis *gredis.Redis, prefix string) gsession.Storage {
	return &RedisStorage{
		StorageRedisHashTable: gsession.NewStorageRedisHashTable(redis, prefix),
	}
}

func (s *RedisStorage) SetSession(ctx context.Context, sessionID string, data *gmap.StrAnyMap, ttl time.Duration) error {
	if ttl == 0 {
		return nil
	}
	return s.StorageRedisHashTable.SetSession(ctx, sessionID, data, ttl)
}

func (s *RedisStorage) UpdateTTL(ctx context.Context, sessionID string, ttl time.Duration) error {
	if ttl == 0 {
		return nil
	}
	return s.StorageRedisHashTable.UpdateTTL(ctx, sessionID, ttl)
}

// Backward-compatible aliases for callers created before session TTLs were
// enabled.
type NoExpireRedisStorage = RedisStorage

func NewNoExpireRedisStorage(redis *gredis.Redis, prefix string) gsession.Storage {
	return NewRedisStorage(redis, prefix)
}
