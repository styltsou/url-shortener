DROP INDEX IF EXISTS idx_links_is_active;

ALTER TABLE links DROP COLUMN IF EXISTS is_active;

