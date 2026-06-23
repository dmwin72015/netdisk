-- name: CreateUser :one
INSERT INTO users (slug, username, email, password_hash, register_method)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserBySlug :one
SELECT * FROM users WHERE slug = $1 LIMIT 1;

-- name: UpdateUserStatus :exec
UPDATE users SET status = $2, updated_at = NOW() WHERE id = $1;
