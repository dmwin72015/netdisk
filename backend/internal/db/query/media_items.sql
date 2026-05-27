-- name: CreateMediaItem :one
INSERT INTO media_items (slug, user_id, user_file_id, physical_file_id, transcode_id, title)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListMediaItemsByUser :many
SELECT mi.*, mt.status AS transcode_status, mt.duration_sec, mt.slug AS transcode_slug
FROM media_items mi
LEFT JOIN media_transcodes mt ON mt.id = mi.transcode_id
WHERE mi.user_id = $1
ORDER BY mi.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountMediaItemsByUser :one
SELECT COUNT(*) FROM media_items WHERE user_id = $1;

-- name: GetMediaItemBySlug :one
SELECT mi.*, mt.status AS transcode_status, mt.duration_sec, mt.hls_dir, mt.slug AS transcode_slug
FROM media_items mi
LEFT JOIN media_transcodes mt ON mt.id = mi.transcode_id
WHERE mi.slug = $1
LIMIT 1;

-- name: GetMediaItemByUserAndFile :one
SELECT * FROM media_items
WHERE user_id = $1 AND user_file_id = $2
LIMIT 1;

-- name: DeleteMediaItem :exec
DELETE FROM media_items WHERE id = $1;
