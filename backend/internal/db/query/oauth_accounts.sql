-- name: CreateOAuthAccount :one
INSERT INTO user_oauth_accounts (user_id, provider, provider_account_id, access_token, refresh_token, token_expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, provider, provider_account_id, access_token, refresh_token, token_expires_at, created_at, updated_at;

-- name: GetOAuthAccountByProviderAndID :one
SELECT id, user_id, provider, provider_account_id, access_token, refresh_token, token_expires_at, created_at, updated_at
FROM user_oauth_accounts
WHERE provider = $1 AND provider_account_id = $2
LIMIT 1;

-- name: GetOAuthAccountsByUserID :many
SELECT id, user_id, provider, provider_account_id, access_token, refresh_token, token_expires_at, created_at, updated_at
FROM user_oauth_accounts
WHERE user_id = $1;
