-- name: CreatePhotoAlbum :one
INSERT INTO photo_albums (slug, user_id, title, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPhotoAlbumBySlug :one
SELECT * FROM photo_albums WHERE slug = $1 AND user_id = $2 LIMIT 1;

-- name: ListPhotoAlbumsByUser :many
SELECT * FROM photo_albums
WHERE user_id = $1
ORDER BY updated_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPhotoAlbumsByUser :one
SELECT COUNT(*) FROM photo_albums WHERE user_id = $1;

-- name: UpdatePhotoAlbum :one
UPDATE photo_albums
SET title = $3, description = $4, cover_file_slug = $5, updated_at = NOW()
WHERE slug = $1 AND user_id = $2
RETURNING *;

-- name: DeletePhotoAlbum :exec
DELETE FROM photo_albums WHERE slug = $1 AND user_id = $2;

-- name: IncrementPhotoAlbumItemCount :exec
UPDATE photo_albums SET item_count = item_count + 1, updated_at = NOW()
WHERE id = $1;

-- name: DecrementPhotoAlbumItemCount :exec
UPDATE photo_albums SET item_count = GREATEST(item_count - 1, 0), updated_at = NOW()
WHERE id = $1;

-- name: AddPhotoToAlbum :one
INSERT INTO photo_album_items (album_id, file_slug, sort_order)
VALUES ($1, $2, $3)
RETURNING *;

-- name: RemovePhotoFromAlbum :exec
DELETE FROM photo_album_items WHERE album_id = $1 AND file_slug = $2;

-- name: ListPhotoAlbumItems :many
SELECT pai.*, uf.file_name, uf.file_size, uf.mime_type, uf.file_category,
       uf.created_at AS file_created_at, pf.file_hash
FROM photo_album_items pai
JOIN user_files uf ON uf.slug = pai.file_slug
LEFT JOIN physical_files pf ON pf.id = uf.physical_file_id
WHERE pai.album_id = $1
ORDER BY pai.sort_order ASC, pai.added_at DESC
LIMIT $2 OFFSET $3;

-- name: CountPhotoAlbumItems :one
SELECT COUNT(*) FROM photo_album_items WHERE album_id = $1;

-- name: GetPhotoAlbumItem :one
SELECT pai.*, uf.file_name, uf.file_size, uf.mime_type, uf.file_category,
       uf.created_at AS file_created_at, pf.file_hash
FROM photo_album_items pai
JOIN user_files uf ON uf.slug = pai.file_slug
LEFT JOIN physical_files pf ON pf.id = uf.physical_file_id
WHERE pai.album_id = $1 AND pai.file_slug = $2
LIMIT 1;

-- name: GetPhotoAlbumItemCountByFile :one
SELECT COUNT(*) FROM photo_album_items WHERE file_slug = $1;

-- name: ListPhotoAlbumsByFile :many
SELECT pa.* FROM photo_albums pa
JOIN photo_album_items pai ON pai.album_id = pa.id
WHERE pai.file_slug = $1 AND pa.user_id = $2
ORDER BY pa.updated_at DESC;
