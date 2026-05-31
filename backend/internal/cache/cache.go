package cache

import (
	"github.com/redis/go-redis/v9"

	"github.com/netdisk/server/internal/config"
)

type Cache struct {
	PreCache      *PreCache
	Challenge     *Challenge
	Chunks        *Chunks
	Lock          *Lock
	MediaProgress *MediaProgress
}

func New(rdb *redis.Client, cfg *config.Config) *Cache {
	return &Cache{
		PreCache:      NewPreCache(rdb, cfg.Cache.PreCacheTTL),
		Challenge:     NewChallenge(rdb, cfg.Cache.ChallengeTTL),
		Chunks:        NewChunks(rdb, cfg.Cache.ChunksTTL),
		Lock:          NewLock(rdb, cfg.Upload.MergeLockTTL),
		MediaProgress: NewMediaProgress(rdb),
	}
}
