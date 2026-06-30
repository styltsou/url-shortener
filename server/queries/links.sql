-- name: TryCreateLink :one
-- sqlc.arg(shortcode) sqlc.arg(original_url) sqlc.arg(user_id) sqlc.narg(expires_at)
INSERT INTO links (shortcode, original_url, user_id, expires_at)
SELECT @shortcode::VARCHAR(20), @original_url::TEXT, @user_id::TEXT, @expires_at
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


-- name: GetLinkByIdAndUserWithTags :one
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
WHERE l.id = $1 
  AND l.user_id = $2 
  AND l.deleted_at IS NULL
GROUP BY l.id;


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
WITH filtered_links AS (
    SELECT l.*
    FROM links l
    WHERE l.user_id = $1
      AND l.deleted_at IS NULL
      AND (
        sqlc.narg('is_active')::boolean IS NULL
        OR (
          sqlc.narg('is_active')::boolean = true
          AND COALESCE(l.is_active, true) = true
          AND (l.expires_at IS NULL OR l.expires_at > NOW())
        )
        OR (
          sqlc.narg('is_active')::boolean = false
          AND (
            COALESCE(l.is_active, true) = false
            OR (l.expires_at IS NOT NULL AND l.expires_at <= NOW())
          )
        )
      )
      AND (
        sqlc.narg('tag_ids')::uuid[] IS NULL
        OR EXISTS (
          SELECT 1
          FROM link_tags filter_lt
          WHERE filter_lt.link_id = l.id
            AND filter_lt.tag_id = ANY(sqlc.narg('tag_ids')::uuid[])
        )
      )
)
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
FROM filtered_links l
LEFT JOIN link_tags lt ON l.id = lt.link_id
LEFT JOIN tags t ON lt.tag_id = t.id
GROUP BY l.id, l.shortcode, l.original_url, l.expires_at, l.is_active, l.created_at, l.updated_at
ORDER BY l.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');


-- name: CountUserLinks :one
WITH filtered_links AS (
    SELECT l.id
    FROM links l
    WHERE l.user_id = $1
      AND l.deleted_at IS NULL
      AND (
        sqlc.narg('is_active')::boolean IS NULL
        OR (
          sqlc.narg('is_active')::boolean = true
          AND COALESCE(l.is_active, true) = true
          AND (l.expires_at IS NULL OR l.expires_at > NOW())
        )
        OR (
          sqlc.narg('is_active')::boolean = false
          AND (
            COALESCE(l.is_active, true) = false
            OR (l.expires_at IS NOT NULL AND l.expires_at <= NOW())
          )
        )
      )
      AND (
        sqlc.narg('tag_ids')::uuid[] IS NULL
        OR EXISTS (
          SELECT 1
          FROM link_tags filter_lt
          WHERE filter_lt.link_id = l.id
            AND filter_lt.tag_id = ANY(sqlc.narg('tag_ids')::uuid[])
        )
      )
)
SELECT COUNT(*) as total FROM filtered_links;


-- name: UpdateLink :one
UPDATE links
SET 
    shortcode = COALESCE(sqlc.narg('shortcode'), shortcode),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    expires_at = CASE
        WHEN sqlc.arg('expires_at_set')::boolean THEN sqlc.narg('expires_at')
        ELSE expires_at
    END,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING id, shortcode, original_url, is_active, expires_at, created_at, updated_at;


-- name: DeleteLink :one
UPDATE links
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING id, shortcode, original_url, is_active, expires_at, created_at, updated_at;


-- name: GetRecentLinks :many
SELECT id, shortcode, original_url, is_active, expires_at, created_at, updated_at
FROM links
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2;


-- name: ListUserLinkIDs :many
SELECT id FROM links
WHERE user_id = $1 AND deleted_at IS NULL;


-- name: CountActiveLinks :one
SELECT count(*) FROM links
WHERE user_id = $1 AND deleted_at IS NULL AND is_active = true
AND (expires_at IS NULL OR expires_at > NOW());
