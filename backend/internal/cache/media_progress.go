package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type MediaProgress struct {
	rdb *redis.Client
}

func NewMediaProgress(rdb *redis.Client) *MediaProgress {
	return &MediaProgress{rdb: rdb}
}

func (mp *MediaProgress) SetProgress(ctx context.Context, transcodeSlug string, pct int) error {
	key := MediaProgressKey(transcodeSlug)
	return mp.rdb.Set(ctx, key, pct, 0).Err()
}

func (mp *MediaProgress) GetProgress(ctx context.Context, transcodeSlug string) (int, error) {
	key := MediaProgressKey(transcodeSlug)
	val, err := mp.rdb.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (mp *MediaProgress) DeleteProgress(ctx context.Context, transcodeSlug string) error {
	key := MediaProgressKey(transcodeSlug)
	return mp.rdb.Del(ctx, key).Err()
}
