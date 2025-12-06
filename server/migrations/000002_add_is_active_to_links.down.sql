-- Drop index
DROP INDEX IF EXISTS idx_links_is_active;

-- Drop column
ALTER TABLE links DROP COLUMN IF EXISTS is_active;

