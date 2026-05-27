-- name: CreateProfile :one
INSERT INTO user_profiles (user_id, display_name)
VALUES ($1, $2)
RETURNING *;

-- name: GetProfileByUserID :one
SELECT * FROM user_profiles WHERE user_id = $1 LIMIT 1;

-- name: UpdateProfile :exec
UPDATE user_profiles
SET display_name = $2, bio = $3, updated_at = NOW()
WHERE user_id = $1;

-- name: UpdateAvatarPath :exec
UPDATE user_profiles
SET avatar_path = $2, updated_at = NOW()
WHERE user_id = $1;
