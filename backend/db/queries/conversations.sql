-- name: ListConversations :many
SELECT * FROM conversations
WHERE user_id = $1
ORDER BY updated_at DESC
LIMIT $2 OFFSET $3;

-- name: GetConversation :one
SELECT * FROM conversations WHERE id = $1 AND user_id = $2;

-- name: CreateConversation :one
INSERT INTO conversations (user_id, title, topic)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateConversationTitle :one
UPDATE conversations SET title = $3, updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteConversation :exec
DELETE FROM conversations WHERE id = $1 AND user_id = $2;

-- name: ListMessages :many
SELECT * FROM messages
WHERE conversation_id = $1
ORDER BY created_at ASC;

-- name: CreateMessage :one
INSERT INTO messages (conversation_id, role, content, input_tokens, output_tokens, model)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
