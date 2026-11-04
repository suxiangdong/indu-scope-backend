package sessionstore

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/os/gsession"
)

// NoExpireRedisStorage uses GoFrame's Redis hash session storage while
// keeping sessions until the application explicitly removes them. The
// underlying hash operations are still handled by GoFrame; only the automatic
// TTL updates are disabled when the server session age is zero.
type NoExpireRedisStorage struct {
	*gsession.StorageRedisHashTable
}

func NewNoExpireRedisStorage(redis *gredis.Redis, prefix string) gsession.Storage {
	return &NoExpireRedisStorage{
		StorageRedisHashTable: gsession.NewStorageRedisHashTable(redis, prefix),
	}
}

func (s *NoExpireRedisStorage) SetSession(ctx context.Context, sessionID string, data *gmap.StrAnyMap, ttl time.Duration) error {
	if ttl == 0 {
		return nil
	}
	return s.StorageRedisHashTable.SetSession(ctx, sessionID, data, ttl)
}

func (s *NoExpireRedisStorage) UpdateTTL(ctx context.Context, sessionID string, ttl time.Duration) error {
	if ttl == 0 {
		return nil
	}
	return s.StorageRedisHashTable.UpdateTTL(ctx, sessionID, ttl)
}
