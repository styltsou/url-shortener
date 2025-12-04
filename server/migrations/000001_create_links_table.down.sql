-- Drop indexes
DROP INDEX IF EXISTS idx_links_deleted_at;
DROP INDEX IF EXISTS idx_links_user_id;
DROP INDEX IF EXISTS idx_links_shortcode;

-- Drop table
DROP TABLE IF EXISTS links;

