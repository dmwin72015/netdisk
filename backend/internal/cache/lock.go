package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Lock struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewLock(rdb *redis.Client, ttl time.Duration) *Lock {
	return &Lock{rdb: rdb, ttl: ttl}
}

func (l *Lock) AcquireMergeLock(ctx context.Context, fileHash string) (bool, error) {
	key := MergeLockKey(fileHash)
	ok, err := l.rdb.SetNX(ctx, key, "1", l.ttl).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (l *Lock) ReleaseMergeLock(ctx context.Context, fileHash string) error {
	key := MergeLockKey(fileHash)
	return l.rdb.Del(ctx, key).Err()
}
