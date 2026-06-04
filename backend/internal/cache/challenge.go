package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Challenge struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewChallenge(rdb *redis.Client, ttl time.Duration) *Challenge {
	return &Challenge{rdb: rdb, ttl: ttl}
}

// SetChallenge stores a challenge for the given user and file hash.
func (c *Challenge) SetChallenge(ctx context.Context, userID int64, fileHash string, offset int, token string) error {
	key := ChallengeKey(userID, fileHash)
	return c.rdb.HSet(ctx, key, map[string]any{
		"offset": fmt.Sprintf("%d", offset),
		"token":  token,
	}).Err()
}

// GetChallenge returns the existing challenge without consuming it.
// Returns (0, "", ErrChallengeExpired) if not found.
func (c *Challenge) GetChallenge(ctx context.Context, userID int64, fileHash string) (int, string, error) {
	key := ChallengeKey(userID, fileHash)
	vals, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return 0, "", fmt.Errorf("get challenge: %w", err)
	}
	if len(vals) == 0 {
		return 0, "", fmt.Errorf("challenge not found")
	}
	offsetStr, ok := vals["offset"]
	if !ok {
		return 0, "", fmt.Errorf("challenge missing offset")
	}
	token, ok := vals["token"]
	if !ok {
		return 0, "", fmt.Errorf("challenge missing token")
	}
	var offset int
	fmt.Sscanf(offsetStr, "%d", &offset)
	return offset, token, nil
}

// ExpireChallenge sets the TTL on a challenge key.
func (c *Challenge) ExpireChallenge(ctx context.Context, userID int64, fileHash string) error {
	key := ChallengeKey(userID, fileHash)
	return c.rdb.Expire(ctx, key, c.ttl).Err()
}

// consumeChallengeScript atomically reads and deletes the challenge.
var consumeChallengeScript = redis.NewScript(`
local v = redis.call('HGETALL', KEYS[1])
if #v == 0 then
	return nil
end
redis.call('DEL', KEYS[1])
return v
`)

// ConsumeChallenge atomically reads and deletes the challenge.
// Returns (offset, token, nil) on success, ("", "", ErrChallengeExpired) if not found.
func (c *Challenge) ConsumeChallenge(ctx context.Context, userID int64, fileHash string) (int, string, error) {
	key := ChallengeKey(userID, fileHash)
	vals, err := consumeChallengeScript.Run(ctx, c.rdb, []string{key}).StringSlice()
	if err == redis.Nil {
		return 0, "", fmt.Errorf("challenge expired or not found")
	}
	if err != nil {
		return 0, "", fmt.Errorf("consume challenge: %w", err)
	}

	// vals is a flat slice: ["offset", "12345", "token", "abc..."]
	var offset int
	var token string
	for i := 0; i < len(vals)-1; i += 2 {
		switch vals[i] {
		case "offset":
			fmt.Sscanf(vals[i+1], "%d", &offset)
		case "token":
			token = vals[i+1]
		}
	}
	return offset, token, nil
}
