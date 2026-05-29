-- Add parent_slug column to user_files for efficient parent directory lookups
ALTER TABLE user_files ADD COLUMN parent_slug VARCHAR(21);

-- Backfill existing rows: set parent_slug from parent's slug
UPDATE user_files f
SET parent_slug = p.slug
FROM user_files p
WHERE f.parent_id = p.id;

-- Create index for efficient parent_slug lookups
CREATE INDEX idx_user_files_parent_slug ON user_files (parent_slug)
    WHERE parent_slug IS NOT NULL;
