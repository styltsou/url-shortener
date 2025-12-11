CREATE TABLE tags (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(30) NOT NULL,
	user_id TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NULL
);

-- Unique index to ensure that a user cannot have duplicate tags
CREATE UNIQUE INDEX index_tags_user_id_name ON tags(user_id, name);