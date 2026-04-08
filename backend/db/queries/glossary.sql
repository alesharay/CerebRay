-- name: ListGlossaryTerms :many
SELECT g.*, n.title as source_note_title
FROM glossary_terms g
LEFT JOIN notes n ON n.id = g.source_note_id
WHERE g.user_id = $1
ORDER BY g.term;

-- name: CreateGlossaryTerm :one
INSERT INTO glossary_terms (user_id, term, definition, source_note_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateGlossaryTerm :one
UPDATE glossary_terms SET
    term = $3,
    definition = $4,
    source_note_id = $5,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteGlossaryTerm :exec
DELETE FROM glossary_terms WHERE id = $1 AND user_id = $2;
