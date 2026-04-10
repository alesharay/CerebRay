-- name: GetFleetingNotesWithAge :many
SELECT n.id, n.title, n.summary, n.core_idea, n.body, n.laymans_terms,
    n.analogy, n.components, n.why_it_matters, n.examples, n.templates,
    n.additional, n.note_type, n.created_at, n.updated_at
FROM notes n
WHERE n.user_id = $1 AND n.status = 'fleeting'
ORDER BY n.created_at ASC;

-- name: GetLifecycleMetrics :many
SELECT
    to_status,
    count(*) as transition_count,
    avg(EXTRACT(EPOCH FROM (ne.created_at - prev.created_at)))::bigint as avg_dwell_seconds
FROM note_events ne
JOIN LATERAL (
    SELECT created_at FROM note_events prev
    WHERE prev.note_id = ne.note_id
        AND prev.created_at < ne.created_at
    ORDER BY prev.created_at DESC
    LIMIT 1
) prev ON true
WHERE ne.user_id = $1
    AND ne.from_status = 'fleeting'
    AND ne.to_status IN ('active', 'sleeping', 'archived')
    AND ne.created_at >= $2
GROUP BY ne.to_status;

-- name: GetConnectionDensity :one
SELECT
    count(DISTINCT n.id)::int as active_notes,
    count(DISTINCT c.id)::int as total_connections,
    count(DISTINCT CASE WHEN c.id IS NULL THEN n.id END)::int as orphan_count
FROM notes n
LEFT JOIN (
    SELECT id, source_id as note_id FROM connections
    UNION ALL
    SELECT id, target_id as note_id FROM connections
) c ON c.note_id = n.id
WHERE n.user_id = $1
    AND n.status IN ('active', 'linked');

-- name: GetNoteTypeDistribution :many
SELECT note_type, count(*) as count
FROM notes
WHERE user_id = $1 AND status IN ('active', 'linked')
GROUP BY note_type;

-- name: GetGlossaryCount :one
SELECT count(*)::int as count FROM glossary_terms WHERE user_id = $1;

-- name: GetSleepingNoteStats :one
SELECT
    count(*)::int as count,
    coalesce(avg(EXTRACT(EPOCH FROM (NOW() - created_at)))::bigint, 0) as avg_age_seconds
FROM notes
WHERE user_id = $1 AND status = 'sleeping';

-- name: GetConversationConversionRate :one
SELECT
    count(DISTINCT conv.id)::int as total_conversations,
    count(DISTINCT n.source_chat_id)::int as conversations_with_notes
FROM conversations conv
LEFT JOIN notes n ON n.source_chat_id = conv.id
WHERE conv.user_id = $1;

-- name: GetStaleNotes :many
SELECT id, title, note_type, updated_at
FROM notes
WHERE user_id = $1
    AND status IN ('active', 'linked')
    AND updated_at < NOW() - make_interval(days => $2::int)
ORDER BY updated_at ASC
LIMIT 10;
