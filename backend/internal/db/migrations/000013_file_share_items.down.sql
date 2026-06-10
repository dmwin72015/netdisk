ALTER TABLE file_shares ADD COLUMN file_id BIGINT;

UPDATE file_shares s
SET file_id = si.file_id
FROM file_share_items si
WHERE si.share_id = s.id;

CREATE INDEX idx_file_shares_file ON file_shares (file_id);

DROP TABLE IF EXISTS file_share_items;
