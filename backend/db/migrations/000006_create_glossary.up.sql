CREATE TABLE glossary_terms (
    id             BIGSERIAL    PRIMARY KEY,
    user_id        BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    term           TEXT         NOT NULL,
    definition     TEXT         NOT NULL,
    source_note_id BIGINT       REFERENCES notes(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, term)
);

CREATE INDEX idx_glossary_user ON glossary_terms (user_id, term);
