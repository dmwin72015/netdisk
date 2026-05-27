-- name: CreateStorageStats :one
INSERT INTO user_storage_stats (user_id, storage_quota)
VALUES ($1, $2)
RETURNING *;

-- name: GetStorageStats :one
SELECT * FROM user_storage_stats WHERE user_id = $1 LIMIT 1;

-- name: AtomicIncrementStorage :one
UPDATE user_storage_stats
SET storage_used = storage_used + $1,
    updated_at = NOW()
WHERE user_id = $2
  AND storage_used + $1 <= storage_quota
RETURNING storage_used, storage_quota;
