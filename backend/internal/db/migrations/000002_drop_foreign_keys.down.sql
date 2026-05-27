-- Re-add all foreign key constraints.

ALTER TABLE user_profiles      ADD CONSTRAINT user_profiles_user_id_fkey           FOREIGN KEY (user_id)          REFERENCES users(id);
ALTER TABLE user_storage_stats ADD CONSTRAINT user_storage_stats_user_id_fkey      FOREIGN KEY (user_id)          REFERENCES users(id);
ALTER TABLE user_levels        ADD CONSTRAINT user_levels_user_id_fkey             FOREIGN KEY (user_id)          REFERENCES users(id);
ALTER TABLE user_transactions  ADD CONSTRAINT user_transactions_user_id_fkey       FOREIGN KEY (user_id)          REFERENCES users(id);
ALTER TABLE user_files         ADD CONSTRAINT user_files_user_id_fkey              FOREIGN KEY (user_id)          REFERENCES users(id);
ALTER TABLE user_files         ADD CONSTRAINT user_files_physical_file_id_fkey     FOREIGN KEY (physical_file_id) REFERENCES physical_files(id);
ALTER TABLE user_files         ADD CONSTRAINT user_files_parent_id_fkey            FOREIGN KEY (parent_id)        REFERENCES user_files(id);
ALTER TABLE upload_tasks       ADD CONSTRAINT upload_tasks_owner_user_id_fkey      FOREIGN KEY (owner_user_id)    REFERENCES users(id);
ALTER TABLE upload_tasks       ADD CONSTRAINT upload_tasks_physical_file_id_fkey   FOREIGN KEY (physical_file_id) REFERENCES physical_files(id);
ALTER TABLE refresh_tokens     ADD CONSTRAINT refresh_tokens_user_id_fkey          FOREIGN KEY (user_id)          REFERENCES users(id);
ALTER TABLE media_transcodes   ADD CONSTRAINT media_transcodes_physical_file_id_fkey FOREIGN KEY (physical_file_id) REFERENCES physical_files(id);
ALTER TABLE media_items        ADD CONSTRAINT media_items_user_id_fkey             FOREIGN KEY (user_id)          REFERENCES users(id);
ALTER TABLE media_items        ADD CONSTRAINT media_items_user_file_id_fkey        FOREIGN KEY (user_file_id)     REFERENCES user_files(id);
ALTER TABLE media_items        ADD CONSTRAINT media_items_physical_file_id_fkey    FOREIGN KEY (physical_file_id) REFERENCES physical_files(id);
ALTER TABLE media_items        ADD CONSTRAINT media_items_transcode_id_fkey        FOREIGN KEY (transcode_id)     REFERENCES media_transcodes(id);
ALTER TABLE media_jobs         ADD CONSTRAINT media_jobs_transcode_id_fkey         FOREIGN KEY (transcode_id)     REFERENCES media_transcodes(id);
