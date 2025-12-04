CREATE TABLE links (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	shortcode VARCHAR(20) NOT NULL,
	original_url TEXT NOT NULL,
	user_id TEXT NOT NULL,
	clicks INTEGER DEFAULT 0,
	expires_at TIMESTAMP DEFAULT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NULL,
	deleted_at TIMESTAMP DEFAULT NULL
);

-- Partial unique index on shortcode (only for non-deleted records)
-- This allows the same shortcode to be reused after deletion
CREATE UNIQUE INDEX idx_links_shortcode ON links(shortcode) WHERE deleted_at IS NULL;

-- Index on user_id for efficient user queries
CREATE INDEX idx_links_user_id ON links(user_id);

-- Index on deleted_at for efficient filtering of deleted records
CREATE INDEX idx_links_deleted_at ON links(deleted_at) WHERE deleted_at IS NOT NULL;
