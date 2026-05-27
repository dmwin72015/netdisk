-- name: CreateTransaction :one
INSERT INTO user_transactions (slug, user_id, transaction_no, type, amount_cents, currency, status, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListTransactionsByUser :many
SELECT * FROM user_transactions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTransactionsByUser :one
SELECT COUNT(*) FROM user_transactions WHERE user_id = $1;
