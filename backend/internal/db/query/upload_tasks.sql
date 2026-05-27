-- name: CreateUploadTask :one
INSERT INTO upload_tasks (slug, owner_user_id, hash_algo, file_hash, pre_hash, file_size, mime_type, total_chunks, chunk_size, status, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetUploadTaskBySlug :one
SELECT * FROM upload_tasks WHERE slug = $1 LIMIT 1;

-- name: GetUploadTaskByHashForUser :one
SELECT * FROM upload_tasks
WHERE owner_user_id = sqlc.arg(owner_user_id) AND file_hash = sqlc.arg(file_hash) AND status IN ('created', 'uploading')
LIMIT 1;

-- name: UpdateUploadTaskStatus :exec
UPDATE upload_tasks
SET status = $2, error_msg = $3, updated_at = NOW()
WHERE id = $1;

-- name: UpdateUploadTaskPhysicalFile :exec
UPDATE upload_tasks
SET physical_file_id = $2, status = 'done', updated_at = NOW()
WHERE id = $1;

-- name: DeleteExpiredTasks :exec
DELETE FROM upload_tasks
WHERE expires_at < NOW() AND status IN ('created', 'uploading', 'merging');
