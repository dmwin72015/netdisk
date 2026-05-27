-- name: CreateTranscode :one
INSERT INTO media_transcodes (slug, physical_file_id, profile, status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTranscodeByPhysicalFileAndProfile :one
SELECT * FROM media_transcodes
WHERE physical_file_id = $1 AND profile = $2
LIMIT 1;

-- name: GetTranscodeBySlug :one
SELECT * FROM media_transcodes WHERE slug = $1 LIMIT 1;

-- name: GetTranscodeByID :one
SELECT * FROM media_transcodes WHERE id = $1 LIMIT 1;

-- name: UpdateTranscodeStatus :exec
UPDATE media_transcodes
SET status = $2, error_msg = $3, updated_at = NOW()
WHERE id = $1;

-- name: UpdateTranscodeHLS :exec
UPDATE media_transcodes
SET hls_dir = $2, duration_sec = $3, status = 'done', updated_at = NOW()
WHERE id = $1;
