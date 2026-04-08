-- name: LogAIUsage :one
INSERT INTO ai_usage_log (user_id, conversation_id, input_tokens, output_tokens, model)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMonthlyUsage :one
SELECT
    coalesce(sum(input_tokens), 0)::int as total_input_tokens,
    coalesce(sum(output_tokens), 0)::int as total_output_tokens,
    coalesce(count(*), 0)::int as total_requests
FROM ai_usage_log
WHERE user_id = $1
    AND created_at >= date_trunc('month', CURRENT_TIMESTAMP)
    AND created_at < date_trunc('month', CURRENT_TIMESTAMP) + interval '1 month';
