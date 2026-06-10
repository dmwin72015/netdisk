CREATE TABLE file_share_items (
    id         BIGSERIAL   PRIMARY KEY,
    share_id   BIGINT      NOT NULL,
    file_id    BIGINT      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(share_id, file_id)
);

INSERT INTO file_share_items (share_id, file_id)
SELECT id, file_id FROM file_shares WHERE file_id IS NOT NULL;

DROP INDEX IF EXISTS idx_file_shares_file;

ALTER TABLE file_shares DROP COLUMN file_id;
