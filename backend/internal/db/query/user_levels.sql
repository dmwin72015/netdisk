-- name: CreateLevel :one
INSERT INTO user_levels (user_id, level_code, level_name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetLevelByUserID :one
SELECT * FROM user_levels WHERE user_id = $1 LIMIT 1;
