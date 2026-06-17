ALTER TABLE upload_tasks ADD COLUMN task_type TEXT NOT NULL DEFAULT 'chunk';
ALTER TABLE upload_tasks ADD COLUMN source_url TEXT;
ALTER TABLE upload_tasks ADD COLUMN received_bytes BIGINT NOT NULL DEFAULT 0;

CREATE INDEX idx_upload_tasks_type_status ON upload_tasks (task_type, status);
