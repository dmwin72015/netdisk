DROP INDEX IF EXISTS idx_upload_tasks_type_status;
ALTER TABLE upload_tasks DROP COLUMN IF EXISTS received_bytes;
ALTER TABLE upload_tasks DROP COLUMN IF EXISTS source_url;
ALTER TABLE upload_tasks DROP COLUMN IF EXISTS task_type;
