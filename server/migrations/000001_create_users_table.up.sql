CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	clerk_id TEXT NOT NULL UNIQUE,
	email TEXT NOT NULL,
	username TEXT DEFAULT NULL,
	avatar_url TEXT DEFAULT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NULL
);

-- fast lookups for when clerk sends webhooks
CREATE INDEX idx_users_clerk_id ON users(clerk_id);