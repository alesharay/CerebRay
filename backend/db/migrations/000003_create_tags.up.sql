CREATE TABLE tags (
    id       BIGSERIAL PRIMARY KEY,
    user_id  BIGINT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name     TEXT      NOT NULL,
    UNIQUE (user_id, name)
);

CREATE TABLE note_tags (
    note_id  BIGINT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    tag_id   BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (note_id, tag_id)
);

CREATE INDEX idx_note_tags_tag ON note_tags (tag_id);
