package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const preCacheTTL = 7 * 24 * time.Hour

type PreCache struct {
	rdb *redis.Client
}

func NewPreCache(rdb *redis.Client) *PreCache {
	return &PreCache{rdb: rdb}
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
	return pc.rdb.Set(ctx, key, physicalFileSlug, preCacheTTL).Err()
}
