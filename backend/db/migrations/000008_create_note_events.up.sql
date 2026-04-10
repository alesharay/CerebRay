CREATE TABLE note_events (
    id          BIGSERIAL    PRIMARY KEY,
    note_id     BIGINT       NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    user_id     BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_status note_status,
    to_status   note_status  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_note_events_note ON note_events (note_id, created_at);
CREATE INDEX idx_note_events_user ON note_events (user_id, created_at);

-- Backfill: create an initial "created" event for every existing note
INSERT INTO note_events (note_id, user_id, from_status, to_status, created_at)
SELECT id, user_id, NULL, status, created_at
FROM notes;
