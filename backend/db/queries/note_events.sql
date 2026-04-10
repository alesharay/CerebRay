-- name: CreateNoteEvent :one
INSERT INTO note_events (note_id, user_id, from_status, to_status)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetNoteCurrentStatus :one
SELECT status FROM notes WHERE id = $1 AND user_id = $2;
