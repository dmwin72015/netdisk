-- Revert storage quota back to 10GB
UPDATE user_storage_stats SET storage_quota = 10737418240 WHERE storage_quota = 536870912000;

-- Revert level back to free
UPDATE user_levels SET level_code = 'free', level_name = '免费用户' WHERE level_code = 'vip1';
