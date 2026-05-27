package cache

import "github.com/redis/go-redis/v9"

type Cache struct {
	PreCache    *PreCache
	Challenge   *Challenge
	Chunks      *Chunks
	Lock        *Lock
	MediaProgress *MediaProgress
}

func New(rdb *redis.Client) *Cache {
	return &Cache{
		PreCache:      NewPreCache(rdb),
		Challenge:     NewChallenge(rdb),
		Chunks:        NewChunks(rdb),
		Lock:          NewLock(rdb),
		MediaProgress: NewMediaProgress(rdb),
	}
}
