-- name: TryCreateLink :one
INSERT INTO links (code, original_url, user_id)
VALUES ($1, $2, $3)
ON CONFLICT (code) DO NOTHING
RETURNING *;

-- name: GetLinkForRedirect :one
SELECT id, original_url, expires_at, clicks 
FROM links
WHERE code = $1
AND (expires_at IS NULL OR expires_at > NOW())
LIMIT 1;

-- name: GetLinkByIdAndUser :one
SELECT * FROM links
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: ListUserLinks :many
SELECT * FROM links
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateLink :one
UPDATE links
SET 
    code = COALESCE(sqlc.narg('code'), code),
    expires_at = COALESCE(sqlc.narg('expires_at'), expires_at),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteLink :exec
DELETE FROM links
WHERE id = $1 AND user_id = $2;

-- name: IncrementClicks :exec
UPDATE links
SET clicks = clicks + 1
WHERE code = $1;
