package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// PreCache maps (fileSize, preHash) → physicalFileSlug so that
// a client presenting the same pre-hash can skip the upload entirely.
type PreCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewPreCache(rdb *redis.Client, ttl time.Duration) *PreCache {
	return &PreCache{rdb: rdb, ttl: ttl}
}

func (pc *PreCache) Get(ctx context.Context, fileSize int64, preHash string) (string, error) {
	key := PreCacheKey(fileSize, preHash)
	val, err := pc.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (pc *PreCache) Set(ctx context.Context, fileSize int64, preHash, physicalFileSlug string) error {
	key := PreCacheKey(fileSize, preHash)
	return pc.rdb.Set(ctx, key, physicalFileSlug, pc.ttl).Err()
}
