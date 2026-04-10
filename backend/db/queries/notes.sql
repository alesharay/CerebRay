-- name: ListNotesByUser :many
SELECT n.*,
    array_agg(DISTINCT t.name) FILTER (WHERE t.name IS NOT NULL) as tag_names,
    count(DISTINCT c.id) as connection_count
FROM notes n
LEFT JOIN note_tags nt ON nt.note_id = n.id
LEFT JOIN tags t ON t.id = nt.tag_id
LEFT JOIN (
    SELECT id, source_id as note_id FROM connections
    UNION ALL
    SELECT id, target_id as note_id FROM connections
) c ON c.note_id = n.id
WHERE n.user_id = $1
GROUP BY n.id
ORDER BY n.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListNotesByStatus :many
SELECT * FROM notes
WHERE user_id = $1 AND status = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetNoteByID :one
SELECT * FROM notes WHERE id = $1 AND user_id = $2;

-- name: CreateNote :one
INSERT INTO notes (
    user_id, title, summary, laymans_terms, analogy, core_idea,
    body, components, why_it_matters, examples, templates, additional,
    note_type, status, tlp, source_chat_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
)
RETURNING *;

-- name: UpdateNote :one
UPDATE notes SET
    title = $3,
    summary = $4,
    laymans_terms = $5,
    analogy = $6,
    core_idea = $7,
    body = $8,
    components = $9,
    why_it_matters = $10,
    examples = $11,
    templates = $12,
    additional = $13,
    note_type = $14,
    status = $15,
    tlp = $16,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: UpdateNoteStatus :one
UPDATE notes SET status = $3, updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteNote :exec
DELETE FROM notes WHERE id = $1 AND user_id = $2;

-- name: SearchNotes :many
SELECT *,
    ts_headline('english', title || ' - ' || summary || ' ' || body,
        plainto_tsquery('english', $2),
        'StartSel=<mark>, StopSel=</mark>, MaxWords=35, MinWords=15'
    ) as snippet
FROM notes
WHERE user_id = $1 AND search_vector @@ plainto_tsquery('english', $2)
ORDER BY ts_rank(search_vector, plainto_tsquery('english', $2)) DESC
LIMIT $3 OFFSET $4;

-- name: ListNotesByTag :many
SELECT n.* FROM notes n
JOIN note_tags nt ON nt.note_id = n.id
JOIN tags t ON t.id = nt.tag_id
WHERE n.user_id = $1 AND t.name = $2
ORDER BY n.created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountNotesByStatus :many
SELECT status, count(*) as count
FROM notes
WHERE user_id = $1
GROUP BY status;

-- name: UpdateNoteSourceChat :one
UPDATE notes SET source_chat_id = $3, updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: GetNoteBySourceChat :one
SELECT * FROM notes WHERE source_chat_id = $1 AND user_id = $2;

-- name: RecentNotes :many
SELECT * FROM notes
WHERE user_id = $1
ORDER BY updated_at DESC
LIMIT $2;
