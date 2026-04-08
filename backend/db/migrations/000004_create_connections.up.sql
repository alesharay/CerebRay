CREATE TABLE connections (
    id          BIGSERIAL    PRIMARY KEY,
    source_id   BIGINT       NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    target_id   BIGINT       NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    label       TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (source_id, target_id),
    CHECK (source_id <> target_id)
);

CREATE INDEX idx_connections_source ON connections (source_id);
CREATE INDEX idx_connections_target ON connections (target_id);
