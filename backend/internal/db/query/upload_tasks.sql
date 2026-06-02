-- name: CreateUploadTask :one
INSERT INTO upload_tasks (slug, owner_user_id, hash_algo, file_hash, pre_hash, file_size, mime_type, original_name, total_chunks, chunk_size, status, expires_at, parent_slug)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetUploadTaskBySlug :one
SELECT * FROM upload_tasks WHERE slug = $1 LIMIT 1;

-- name: GetUploadTaskByHashForUser :one
SELECT * FROM upload_tasks
WHERE owner_user_id = sqlc.arg(owner_user_id) AND file_hash = sqlc.arg(file_hash) AND status IN ('created', 'uploading')
LIMIT 1;

-- name: ListUploadTasksByUser :many
SELECT * FROM upload_tasks
WHERE owner_user_id = $1
  AND (sqlc.narg(start_date)::timestamptz IS NULL OR created_at >= sqlc.narg(start_date)::timestamptz)
  AND (sqlc.narg(end_date)::timestamptz IS NULL OR created_at <= sqlc.narg(end_date)::timestamptz)
  AND (sqlc.narg(status)::varchar IS NULL OR status = sqlc.narg(status)::varchar)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUploadTasksByUser :one
SELECT COUNT(*) FROM upload_tasks
WHERE owner_user_id = $1
  AND (sqlc.narg(start_date)::timestamptz IS NULL OR created_at >= sqlc.narg(start_date)::timestamptz)
  AND (sqlc.narg(end_date)::timestamptz IS NULL OR created_at <= sqlc.narg(end_date)::timestamptz)
  AND (sqlc.narg(status)::varchar IS NULL OR status = sqlc.narg(status)::varchar);

-- name: UpdateUploadTaskStatus :exec
UPDATE upload_tasks
SET status = $2, error_msg = $3, updated_at = NOW()
WHERE id = $1;

-- name: UpdateUploadTaskPhysicalFile :exec
UPDATE upload_tasks
SET physical_file_id = $2, status = 'done', updated_at = NOW()
WHERE id = $1;

-- name: UpdateUploadTaskFileHash :exec
UPDATE upload_tasks
SET file_hash = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateUploadTaskPreHash :exec
UPDATE upload_tasks
SET pre_hash = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteExpiredTasks :exec
DELETE FROM upload_tasks
WHERE expires_at < NOW() AND status IN ('created', 'uploading', 'merging');

-- name: DeleteUploadTaskBySlug :exec
DELETE FROM upload_tasks WHERE slug = $1 AND owner_user_id = $2;

-- name: DeleteUploadTasksBySlugs :exec
DELETE FROM upload_tasks WHERE slug = ANY($1::varchar[]) AND owner_user_id = $2;
