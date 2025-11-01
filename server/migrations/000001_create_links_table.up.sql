CREATE TABLE links (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	code VARCHAR(20) NOT NULL UNIQUE,
	original_url TEXT NOT NULL,
	user_id TEXT NOT NULL,
	clicks INTEGER DEFAULT 0,
	expires_at TIMESTAMP DEFAULT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_links_code ON links(code);
CREATE INDEX idx_links_user_id ON links(user_id)