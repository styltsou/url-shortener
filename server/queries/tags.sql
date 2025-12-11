-- name: ListUserTags :many
SELECT id, name, created_at, updated_at FROM tags
WHERE user_id = $1
ORDER BY name;

-- name: CreateTag :one
INSERT INTO tags (name, user_id)
VALUES ($1, $2)
RETURNING id, name, created_at, updated_at;

-- name: UpdateTag :one
UPDATE tags
SET 
	name = $1, 
	updated_at = NOW()
WHERE id = $2 AND user_id = $3
RETURNING id, name, created_at, updated_at;

-- name: DeleteTag :one
DELETE FROM tags
WHERE id = $1 AND user_id = $2
RETURNING id, name, created_at, updated_at;

-- name: DeleteTags :many
DELETE FROM tags
WHERE id = ANY($1::uuid[]) AND user_id = $2
RETURNING id, name, created_at, updated_at;