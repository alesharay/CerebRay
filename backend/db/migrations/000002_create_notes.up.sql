CREATE TYPE note_type AS ENUM (
    'concept', 'theory', 'insight', 'quote',
    'reference', 'question', 'structure', 'guide'
);

CREATE TYPE note_status AS ENUM (
    'fleeting', 'sleeping', 'active', 'linked', 'archived'
);

CREATE TYPE note_tlp AS ENUM ('clear', 'green', 'amber', 'red');

CREATE TABLE notes (
    id              BIGSERIAL    PRIMARY KEY,
    user_id         BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title           TEXT         NOT NULL,
    summary         TEXT         NOT NULL DEFAULT '',
    laymans_terms   TEXT         NOT NULL DEFAULT '',
    analogy         TEXT         NOT NULL DEFAULT '',
    core_idea       TEXT         NOT NULL DEFAULT '',
    body            TEXT         NOT NULL DEFAULT '',
    components      TEXT         NOT NULL DEFAULT '',
    why_it_matters  TEXT         NOT NULL DEFAULT '',
    examples        TEXT         NOT NULL DEFAULT '',
    templates       TEXT         NOT NULL DEFAULT '',
    additional      TEXT         NOT NULL DEFAULT '',
    note_type       note_type    NOT NULL DEFAULT 'concept',
    status          note_status  NOT NULL DEFAULT 'fleeting',
    tlp             note_tlp     NOT NULL DEFAULT 'clear',
    source_chat_id  BIGINT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    search_vector   tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(summary, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(core_idea, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(body, '')), 'C')
    ) STORED
);

CREATE INDEX idx_notes_user_id ON notes (user_id);
CREATE INDEX idx_notes_status ON notes (user_id, status);
CREATE INDEX idx_notes_type ON notes (user_id, note_type);
CREATE INDEX idx_notes_search ON notes USING GIN (search_vector);
CREATE INDEX idx_notes_created ON notes (user_id, created_at DESC);
