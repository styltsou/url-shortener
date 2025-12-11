-- name: TryCreateLink :one
-- sqlc.arg(shortcode) sqlc.arg(original_url) sqlc.arg(user_id)
INSERT INTO links (shortcode, original_url, user_id)
SELECT @shortcode::VARCHAR(20), @original_url::TEXT, @user_id::TEXT
WHERE NOT EXISTS (
    SELECT 1 FROM links 
    WHERE shortcode = @shortcode::VARCHAR(20) AND deleted_at IS NULL
)
RETURNING id, shortcode, original_url, expires_at, is_active, created_at, updated_at;


-- name: GetLinkForRedirect :one
SELECT id, original_url
FROM links
WHERE shortcode = $1
AND deleted_at IS NULL
AND is_active = true
AND (expires_at IS NULL OR expires_at > NOW())
LIMIT 1;


-- name: GetLinkByIdAndUser :one
SELECT id, shortcode, original_url, expires_at, is_active, created_at, updated_at
FROM links
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
LIMIT 1;


-- name: GetLinkByShortcodeAndUser :one
SELECT 
    l.id,
    l.shortcode,
    l.original_url,
    l.expires_at,
    l.is_active,
    l.created_at,
    l.updated_at,
    COALESCE(
        json_agg(
            json_build_object(
                'id', t.id,
                'name', t.name,
                'created_at', t.created_at
            )
        ) FILTER (WHERE t.id IS NOT NULL),
        '[]'::json
    ) as tags
FROM links l
LEFT JOIN link_tags lt ON l.id = lt.link_id
LEFT JOIN tags t ON lt.tag_id = t.id
WHERE l.shortcode = $1 
  AND l.user_id = $2 
  AND l.deleted_at IS NULL
GROUP BY l.id;


-- name: ListUserLinks :many
SELECT 
    l.id,
    l.shortcode,
    l.original_url,
    l.expires_at,
    l.is_active,
    l.created_at,
    l.updated_at,
    COALESCE(
        json_agg(
            json_build_object(
                'id', t.id,
                'name', t.name,
                'created_at', t.created_at
            )
        ) FILTER (WHERE t.id IS NOT NULL),
        '[]'::json
    ) as tags
FROM links l
LEFT JOIN link_tags lt ON l.id = lt.link_id
LEFT JOIN tags t ON lt.tag_id = t.id
WHERE l.user_id = $1 
  AND l.deleted_at IS NULL
GROUP BY l.id
ORDER BY l.created_at DESC;


-- name: UpdateLink :one
UPDATE links
SET 
    shortcode = COALESCE(sqlc.narg('shortcode'), shortcode),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    expires_at = COALESCE(sqlc.narg('expires_at'), expires_at),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING id, shortcode, original_url, is_active, expires_at, created_at, updated_at;


-- name: DeleteLink :one
UPDATE links
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING id, shortcode, original_url, is_active, expires_at, created_at, updated_at;
