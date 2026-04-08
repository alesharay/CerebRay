-- name: ListConnectionsForNote :many
SELECT c.*,
    CASE WHEN c.source_id = $1 THEN n2.title ELSE n1.title END as connected_title,
    CASE WHEN c.source_id = $1 THEN c.target_id ELSE c.source_id END as connected_id
FROM connections c
JOIN notes n1 ON n1.id = c.source_id
JOIN notes n2 ON n2.id = c.target_id
WHERE c.source_id = $1 OR c.target_id = $1;

-- name: CreateConnection :one
INSERT INTO connections (source_id, target_id, label)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteConnection :exec
DELETE FROM connections WHERE id = $1;

-- name: GetGraphData :many
SELECT c.source_id, c.target_id, c.label,
    n1.title as source_title, n1.note_type as source_type,
    n2.title as target_title, n2.note_type as target_type
FROM connections c
JOIN notes n1 ON n1.id = c.source_id
JOIN notes n2 ON n2.id = c.target_id
WHERE n1.user_id = $1;
