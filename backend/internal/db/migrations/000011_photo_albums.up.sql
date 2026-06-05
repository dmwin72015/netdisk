CREATE TABLE IF NOT EXISTS photo_albums (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(21) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    title VARCHAR(256) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    cover_file_slug VARCHAR(21),
    item_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_photo_albums_user ON photo_albums (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS photo_album_items (
    id BIGSERIAL PRIMARY KEY,
    album_id BIGINT NOT NULL REFERENCES photo_albums(id) ON DELETE CASCADE,
    file_slug VARCHAR(21) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(album_id, file_slug)
);

CREATE INDEX IF NOT EXISTS idx_photo_album_items_album ON photo_album_items (album_id);
CREATE INDEX IF NOT EXISTS idx_photo_album_items_file ON photo_album_items (file_slug);

ALTER TABLE photo_albums ADD CONSTRAINT fk_photo_albums_cover
    FOREIGN KEY (cover_file_slug) REFERENCES user_files(slug) ON DELETE SET NULL;
