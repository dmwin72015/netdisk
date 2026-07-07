-- name: CreateActivityLog :exec
INSERT INTO user_activity_logs (user_id, action, resource_type, resource_name,
    ip, ip_region, user_agent, os, browser, device_id, extra)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: ListActivityLogsByUser :many
SELECT id, user_id, action, resource_type, resource_name, ip, ip_region, user_agent, os, browser, device_id, extra, created_at FROM user_activity_logs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListSecurityLogsByUser :many
SELECT id, user_id, action, resource_type, resource_name, ip, ip_region, user_agent, os, browser, device_id, extra, created_at FROM user_activity_logs
WHERE user_id = $1
  AND action IN ('user.login', 'user.register', 'user.logout',
                 'user.oauth_login', 'user.password_change')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountActivityLogsByUser :one
SELECT COUNT(*) FROM user_activity_logs WHERE user_id = $1;

-- name: CountSecurityLogsByUser :one
SELECT COUNT(*) FROM user_activity_logs
WHERE user_id = $1
  AND action IN ('user.login', 'user.register', 'user.logout',
                 'user.oauth_login', 'user.password_change');
