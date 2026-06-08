-- name: CreateFile :one
INSERT INTO user_files (
    slug,
    user_id,
    physical_file_id,
    parent_id,
    parent_slug,
    file_name,
    is_dir,
    file_size,
    mime_type,
    file_category,
    is_system,
    system_kind
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetFileBySlug :one
SELECT * FROM user_files WHERE slug = $1 LIMIT 1;

-- name: GetFileBySlugForUser :one
SELECT * FROM user_files WHERE slug = $1 AND user_id = $2 LIMIT 1;

-- name: GetFileByID :one
SELECT * FROM user_files WHERE id = $1 LIMIT 1;

-- name: RenameFile :exec
UPDATE user_files
SET file_name = $2, updated_at = NOW()
WHERE id = $1;

-- name: MoveFile :exec
UPDATE user_files
SET parent_id = $2, parent_slug = $3, updated_at = NOW()
WHERE id = $1;

-- name: SetStarred :exec
UPDATE user_files
SET is_starred = $2, updated_at = NOW()
WHERE id = $1;

-- name: SetTrashed :exec
UPDATE user_files
SET is_trashed = $2, trashed_at = CASE WHEN $2 THEN NOW() ELSE NULL END, updated_at = NOW()
WHERE id = $1;

-- name: RestoreFile :exec
UPDATE user_files
SET is_trashed = FALSE, trashed_at = NULL, updated_at = NOW()
WHERE id = $1;

-- name: DeleteFile :exec
DELETE FROM user_files WHERE id = $1;

-- name: CheckNameConflict :one
SELECT id, slug, file_name, is_dir, file_size FROM user_files
WHERE user_id = sqlc.arg(user_id)
  AND is_trashed = FALSE
  AND file_name = sqlc.arg(file_name)
  AND is_system = sqlc.arg(is_system)
  AND (
    (sqlc.narg(parent_id)::bigint IS NULL AND parent_id IS NULL)
    OR
    (sqlc.narg(parent_id)::bigint IS NOT NULL AND parent_id = sqlc.narg(parent_id)::bigint)
  )
LIMIT 1;

-- name: GetSystemDirByKind :one
SELECT * FROM user_files
WHERE user_id = sqlc.arg(user_id)
  AND is_trashed = FALSE
  AND is_dir = TRUE
  AND is_system = TRUE
  AND system_kind = sqlc.arg(system_kind)
  AND (
    (sqlc.narg(parent_id)::bigint IS NULL AND parent_id IS NULL)
    OR
    (sqlc.narg(parent_id)::bigint IS NOT NULL AND parent_id = sqlc.narg(parent_id)::bigint)
  )
LIMIT 1;

-- name: GetAncestors :many
WITH RECURSIVE ancestors AS (
    SELECT uf.id, uf.slug, uf.parent_id, uf.file_name, uf.is_dir
    FROM user_files uf WHERE uf.id = $1
    UNION ALL
    SELECT f.id, f.slug, f.parent_id, f.file_name, f.is_dir
    FROM user_files f
    JOIN ancestors a ON f.id = a.parent_id
)
SELECT a.id, a.slug, a.parent_id, a.file_name, a.is_dir FROM ancestors a ORDER BY a.id ASC;

-- name: GetUserFilesByPhysicalFileID :many
SELECT file_name, parent_slug FROM user_files
WHERE physical_file_id = $1 AND user_id = $2 AND is_trashed = FALSE
LIMIT 5;

-- name: GetActiveFileByMediaUploadIdentity :one
SELECT * FROM user_files
WHERE user_id = sqlc.arg(user_id)
  AND parent_id = sqlc.arg(parent_id)
  AND physical_file_id = sqlc.arg(physical_file_id)
  AND file_name = sqlc.arg(file_name)
  AND is_dir = FALSE
  AND is_trashed = FALSE
  AND is_system = FALSE
LIMIT 1;

-- name: GetExpiredTrashedFiles :many
SELECT uf.id, uf.user_id, uf.is_dir, uf.physical_file_id, uf.file_size
FROM user_files uf
WHERE uf.is_trashed = TRUE
  AND uf.is_system = FALSE
  AND uf.trashed_at < NOW() - INTERVAL '30 days';
