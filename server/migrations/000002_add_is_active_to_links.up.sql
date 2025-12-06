-- Add is_active column to links table
ALTER TABLE links ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

-- Index on is_active for efficient filtering of active/inactive links
CREATE INDEX idx_links_is_active ON links(is_active) WHERE is_active = true;

