-- Update existing users' storage quota from 10GB to 500GB
UPDATE user_storage_stats SET storage_quota = 536870912000 WHERE storage_quota = 10737418240;

-- Update existing users' level from free to vip1
UPDATE user_levels SET level_code = 'vip1', level_name = 'VIP1' WHERE level_code = 'free';
