-- name: AddTagsToLink :exec
-- Adds multiple tags to a link, ensuring both link and tags belong to the same user
INSERT INTO link_tags (link_id, tag_id)
SELECT $1, unnest(sqlc.arg(tag_i_ds)::uuid[])
WHERE EXISTS (
    SELECT 1 FROM links l
    WHERE l.id = $1 AND l.user_id = $2 AND l.deleted_at IS NULL
)
AND EXISTS (
    SELECT 1 FROM tags t
    WHERE t.id = ANY(sqlc.arg(tag_i_ds)::uuid[]) AND t.user_id = $2
)
ON CONFLICT (link_id, tag_id) DO NOTHING;

-- name: RemoveTagsFromLink :exec
-- Removes multiple tags from a link, ensuring both link and tags belong to the same user
DELETE FROM link_tags
WHERE link_id = $1 
  AND tag_id = ANY(sqlc.arg(tag_i_ds)::uuid[])
  AND EXISTS (
      SELECT 1 FROM links l
      WHERE l.id = $1 AND l.user_id = $2 AND l.deleted_at IS NULL
  )
  AND EXISTS (
      SELECT 1 FROM tags t
      WHERE t.id = ANY(sqlc.arg(tag_i_ds)::uuid[]) AND t.user_id = $2
  );

