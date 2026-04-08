-- name: ListTagsByUser :many
SELECT t.id, t.name, count(nt.note_id) as note_count
FROM tags t
LEFT JOIN note_tags nt ON nt.tag_id = t.id
WHERE t.user_id = $1
GROUP BY t.id, t.name
ORDER BY t.name;

-- name: CreateTag :one
INSERT INTO tags (user_id, name) VALUES ($1, $2)
ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
RETURNING *;

-- name: AddNoteTag :exec
INSERT INTO note_tags (note_id, tag_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveNoteTag :exec
DELETE FROM note_tags WHERE note_id = $1 AND tag_id = $2;

-- name: GetTagsForNote :many
SELECT t.* FROM tags t
JOIN note_tags nt ON nt.tag_id = t.id
WHERE nt.note_id = $1;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = $1 AND user_id = $2;
