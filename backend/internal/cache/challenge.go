package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const challengeTTL = 2 * time.Minute

type Challenge struct {
	rdb *redis.Client
}

func NewChallenge(rdb *redis.Client) *Challenge {
	return &Challenge{rdb: rdb}
}

// SetChallenge stores a challenge for the given user and file hash.
func (c *Challenge) SetChallenge(ctx context.Context, userID int64, fileHash string, offset int, token string) error {
	key := ChallengeKey(userID, fileHash)
	return c.rdb.HSet(ctx, key, map[string]any{
		"offset": fmt.Sprintf("%d", offset),
		"token":  token,
	}).Err()
}

// ExpireChallenge sets the TTL on a challenge key.
func (c *Challenge) ExpireChallenge(ctx context.Context, userID int64, fileHash string) error {
	key := ChallengeKey(userID, fileHash)
	return c.rdb.Expire(ctx, key, challengeTTL).Err()
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
