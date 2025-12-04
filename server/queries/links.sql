-- name: TryCreateLink :one
-- sqlc.arg(shortcode) sqlc.arg(original_url) sqlc.arg(user_id)
INSERT INTO links (shortcode, original_url, user_id)
SELECT @shortcode::VARCHAR(20), @original_url::TEXT, @user_id::TEXT
WHERE NOT EXISTS (
    SELECT 1 FROM links 
    WHERE shortcode = @shortcode::VARCHAR(20) AND deleted_at IS NULL
)
RETURNING *;

-- name: GetLinkForRedirect :one
SELECT id, original_url, expires_at, clicks 
FROM links
WHERE shortcode = $1
AND deleted_at IS NULL
AND (expires_at IS NULL OR expires_at > NOW())
LIMIT 1;

-- name: GetLinkByIdAndUser :one
SELECT * FROM links
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: ListUserLinks :many
SELECT * FROM links
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: UpdateLink :one
UPDATE links
SET 
    shortcode = COALESCE(sqlc.narg('shortcode'), shortcode),
    expires_at = COALESCE(sqlc.narg('expires_at'), expires_at),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteLink :execrows
UPDATE links
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: IncrementClicks :exec
UPDATE links
SET clicks = clicks + 1
WHERE shortcode = $1 AND deleted_at IS NULL;
