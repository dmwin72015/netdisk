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
