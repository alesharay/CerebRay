-- name: ListConnectionsForNote :many
SELECT c.*,
    CASE WHEN c.source_id = $1 THEN n2.title ELSE n1.title END as connected_title,
    CASE WHEN c.source_id = $1 THEN c.target_id ELSE c.source_id END as connected_id,
    CASE WHEN c.source_id = $1 THEN 'outgoing' ELSE 'incoming' END as direction
FROM connections c
JOIN notes n1 ON n1.id = c.source_id
JOIN notes n2 ON n2.id = c.target_id
WHERE (c.source_id = $1 OR c.target_id = $1)
  AND n1.user_id = $2
  AND n2.user_id = $2;

-- name: CreateConnection :one
INSERT INTO connections (source_id, target_id, label)
SELECT sqlc.arg(source_id)::bigint, sqlc.arg(target_id)::bigint, sqlc.arg(label)::text
WHERE EXISTS (SELECT 1 FROM notes src WHERE src.id = sqlc.arg(source_id) AND src.user_id = sqlc.arg(user_id))
  AND EXISTS (SELECT 1 FROM notes tgt WHERE tgt.id = sqlc.arg(target_id) AND tgt.user_id = sqlc.arg(user_id))
RETURNING *;

-- name: DeleteConnection :exec
DELETE FROM connections c
USING notes n
WHERE c.id = $1 AND n.id = c.source_id AND n.user_id = $2;

-- name: GetGraphData :many
SELECT c.source_id, c.target_id, c.label,
    n1.title as source_title, n1.note_type as source_type,
    n2.title as target_title, n2.note_type as target_type
FROM connections c
JOIN notes n1 ON n1.id = c.source_id
JOIN notes n2 ON n2.id = c.target_id
WHERE n1.user_id = $1;
