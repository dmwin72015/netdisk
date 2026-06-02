package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Chunks stores the set of uploaded chunk indices for an upload session.
type Chunks struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewChunks(rdb *redis.Client, ttl time.Duration) *Chunks {
	return &Chunks{rdb: rdb, ttl: ttl}
}

func (ch *Chunks) AddChunk(ctx context.Context, uploadSlug string, chunkIndex int) error {
	key := ChunksKey(uploadSlug)
	pipe := ch.rdb.Pipeline()
	pipe.SAdd(ctx, key, chunkIndex)
	pipe.Expire(ctx, key, ch.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (ch *Chunks) ListChunks(ctx context.Context, uploadSlug string) ([]int, error) {
	key := ChunksKey(uploadSlug)
	members, err := ch.rdb.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	indices := make([]int, 0, len(members))
	for _, m := range members {
		idx, err := strconv.Atoi(m)
		if err != nil {
			continue
		}
		indices = append(indices, idx)
	}
	return indices, nil
}

func (ch *Chunks) DeleteChunks(ctx context.Context, uploadSlug string) error {
	key := ChunksKey(uploadSlug)
	return ch.rdb.Del(ctx, key).Err()
}

func (ch *Chunks) ChunkCount(ctx context.Context, uploadSlug string) (int64, error) {
	key := ChunksKey(uploadSlug)
	count, err := ch.rdb.SCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("chunk count: %w", err)
	}
	return count, nil
}

func (ch *Chunks) HasChunk(ctx context.Context, uploadSlug string, chunkIndex int) (bool, error) {
	key := ChunksKey(uploadSlug)
	return ch.rdb.SIsMember(ctx, key, chunkIndex).Result()
}
