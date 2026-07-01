-- name: CreatePhysicalFile :one
INSERT INTO physical_files (slug, hash_algo, file_hash, pre_hash, file_size, mime_type, storage_path, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (hash_algo, file_hash) DO NOTHING
RETURNING *;

-- name: GetPhysicalFileByHash :one
SELECT * FROM physical_files
WHERE hash_algo = $1 AND file_hash = $2 AND status = 'completed'
LIMIT 1;

-- name: GetPhysicalFileBySlug :one
SELECT * FROM physical_files WHERE slug = $1 LIMIT 1;

-- name: GetPhysicalFileByPreHash :one
SELECT * FROM physical_files
WHERE file_size = $1 AND pre_hash = $2 AND status = 'completed'
LIMIT 1;

-- name: GetPhysicalFileByID :one
SELECT * FROM physical_files WHERE id = $1 LIMIT 1;

-- name: DeletePhysicalFile :exec
DELETE FROM physical_files WHERE id = $1;

-- name: CountReferencesByFileID :one
SELECT
    (SELECT COUNT(*) FROM user_files uf WHERE uf.physical_file_id = $1)
    +
    (SELECT COUNT(*) FROM media_items mi WHERE mi.physical_file_id = $1);

-- name: ListPhysicalFiles :many
SELECT pf.*,
    (SELECT COUNT(*) FROM user_files uf WHERE uf.physical_file_id = pf.id) AS user_file_count,
    (SELECT COUNT(*) FROM media_items mi WHERE mi.physical_file_id = pf.id) AS media_item_count
FROM physical_files pf
WHERE (sqlc.narg('status')::text IS NULL OR pf.status = sqlc.narg('status'))
  AND (sqlc.narg('search')::text IS NULL
       OR pf.slug ILIKE '%' || sqlc.narg('search')::text || '%'
       OR pf.file_hash ILIKE '%' || sqlc.narg('search')::text || '%'
       OR pf.storage_path ILIKE '%' || sqlc.narg('search')::text || '%')
  AND (sqlc.narg('hash_filter')::text IS NULL
       OR pf.file_hash ILIKE '%' || sqlc.narg('hash_filter')::text || '%')
  AND (sqlc.narg('mime_filter')::text IS NULL
       OR pf.mime_type ILIKE '%' || sqlc.narg('mime_filter')::text || '%')
  AND (sqlc.narg('min_size')::bigint IS NULL OR pf.file_size >= sqlc.narg('min_size')::bigint)
  AND (sqlc.narg('max_size')::bigint IS NULL OR pf.file_size <= sqlc.narg('max_size')::bigint)
  AND (sqlc.narg('created_from')::timestamptz IS NULL
       OR pf.created_at >= sqlc.narg('created_from')::timestamptz)
  AND (sqlc.narg('created_to')::timestamptz IS NULL
       OR pf.created_at <= sqlc.narg('created_to')::timestamptz)
ORDER BY pf.created_at DESC
LIMIT sqlc.arg('lim')
OFFSET sqlc.arg('off');

-- name: CountPhysicalFiles :one
SELECT COUNT(*) FROM physical_files pf
WHERE (sqlc.narg('status')::text IS NULL OR pf.status = sqlc.narg('status'))
  AND (sqlc.narg('search')::text IS NULL
       OR pf.slug ILIKE '%' || sqlc.narg('search')::text || '%'
       OR pf.file_hash ILIKE '%' || sqlc.narg('search')::text || '%'
       OR pf.storage_path ILIKE '%' || sqlc.narg('search')::text || '%')
  AND (sqlc.narg('hash_filter')::text IS NULL
       OR pf.file_hash ILIKE '%' || sqlc.narg('hash_filter')::text || '%')
  AND (sqlc.narg('mime_filter')::text IS NULL
       OR pf.mime_type ILIKE '%' || sqlc.narg('mime_filter')::text || '%')
  AND (sqlc.narg('min_size')::bigint IS NULL OR pf.file_size >= sqlc.narg('min_size')::bigint)
  AND (sqlc.narg('max_size')::bigint IS NULL OR pf.file_size <= sqlc.narg('max_size')::bigint)
  AND (sqlc.narg('created_from')::timestamptz IS NULL
       OR pf.created_at >= sqlc.narg('created_from')::timestamptz)
  AND (sqlc.narg('created_to')::timestamptz IS NULL
       OR pf.created_at <= sqlc.narg('created_to')::timestamptz);
