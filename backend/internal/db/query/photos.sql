-- name: ListImageFiles :many
SELECT uf.slug, uf.file_name, uf.file_size, uf.mime_type, uf.file_category,
       uf.is_starred, uf.created_at, uf.updated_at,
       pf.file_hash
FROM user_files uf
LEFT JOIN physical_files pf ON pf.id = uf.physical_file_id
WHERE uf.user_id = $1
  AND uf.is_dir = FALSE
  AND uf.is_trashed = FALSE
  AND uf.file_category = 'image'
ORDER BY uf.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountImageFiles :one
SELECT COUNT(*) FROM user_files
WHERE user_id = $1
  AND is_dir = FALSE
  AND is_trashed = FALSE
  AND file_category = 'image';
