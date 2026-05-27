-- name: CreateMediaJob :one
INSERT INTO media_jobs (slug, transcode_id, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPendingJobs :many
SELECT mj.*
FROM media_jobs mj
JOIN media_transcodes mt ON mt.id = mj.transcode_id
WHERE mj.status = 'pending'
  AND mt.status = 'pending'
ORDER BY mj.created_at
LIMIT $1
FOR UPDATE OF mj SKIP LOCKED;

-- name: UpdateJobStatus :exec
UPDATE media_jobs
SET status = $2, error_msg = $3, updated_at = NOW()
WHERE id = $1;

-- name: GetJobBySlug :one
SELECT * FROM media_jobs WHERE slug = $1 LIMIT 1;
