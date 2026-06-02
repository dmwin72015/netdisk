-- Drop all foreign key constraints. Referential integrity is enforced at the application layer.

ALTER TABLE user_profiles      DROP CONSTRAINT IF EXISTS user_profiles_user_id_fkey;
ALTER TABLE user_storage_stats DROP CONSTRAINT IF EXISTS user_storage_stats_user_id_fkey;
ALTER TABLE user_levels        DROP CONSTRAINT IF EXISTS user_levels_user_id_fkey;
ALTER TABLE user_transactions  DROP CONSTRAINT IF EXISTS user_transactions_user_id_fkey;
ALTER TABLE user_files         DROP CONSTRAINT IF EXISTS user_files_user_id_fkey;
ALTER TABLE user_files         DROP CONSTRAINT IF EXISTS user_files_physical_file_id_fkey;
ALTER TABLE user_files         DROP CONSTRAINT IF EXISTS user_files_parent_id_fkey;
ALTER TABLE upload_tasks       DROP CONSTRAINT IF EXISTS upload_tasks_owner_user_id_fkey;
ALTER TABLE upload_tasks       DROP CONSTRAINT IF EXISTS upload_tasks_physical_file_id_fkey;
ALTER TABLE refresh_tokens     DROP CONSTRAINT IF EXISTS refresh_tokens_user_id_fkey;
ALTER TABLE media_transcodes   DROP CONSTRAINT IF EXISTS media_transcodes_physical_file_id_fkey;
ALTER TABLE media_items        DROP CONSTRAINT IF EXISTS media_items_user_id_fkey;
ALTER TABLE media_items        DROP CONSTRAINT IF EXISTS media_items_user_file_id_fkey;
ALTER TABLE media_items        DROP CONSTRAINT IF EXISTS media_items_physical_file_id_fkey;
ALTER TABLE media_items        DROP CONSTRAINT IF EXISTS media_items_transcode_id_fkey;
ALTER TABLE media_jobs         DROP CONSTRAINT IF EXISTS media_jobs_transcode_id_fkey;
