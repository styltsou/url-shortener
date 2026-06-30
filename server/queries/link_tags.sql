-- name: AddTagsToLink :exec
-- Adds owned tags to an owned link
INSERT INTO link_tags (link_id, tag_id)
SELECT $1, t.id
FROM tags t
JOIN links l ON l.id = $1 AND l.user_id = $2 AND l.deleted_at IS NULL
WHERE t.id = ANY(sqlc.arg(tag_i_ds)::uuid[]) AND t.user_id = $2
ON CONFLICT (link_id, tag_id) DO NOTHING;

-- name: RemoveTagsFromLink :exec
-- Removes owned tags from an owned link
DELETE FROM link_tags lt
USING links l, tags t
WHERE lt.link_id = $1
  AND lt.tag_id = t.id
  AND l.id = lt.link_id
  AND l.user_id = $2
  AND l.deleted_at IS NULL
  AND t.user_id = $2
  AND t.id = ANY(sqlc.arg(tag_i_ds)::uuid[]);
