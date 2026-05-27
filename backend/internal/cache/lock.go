package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const mergeLockTTL = 10 * time.Minute

type Lock struct {
	rdb *redis.Client
}

func NewLock(rdb *redis.Client) *Lock {
	return &Lock{rdb: rdb}
}

func (l *Lock) AcquireMergeLock(ctx context.Context, fileHash string) (bool, error) {
	key := MergeLockKey(fileHash)
	ok, err := l.rdb.SetNX(ctx, key, "1", mergeLockTTL).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (l *Lock) ReleaseMergeLock(ctx context.Context, fileHash string) error {
	key := MergeLockKey(fileHash)
	return l.rdb.Del(ctx, key).Err()
}
