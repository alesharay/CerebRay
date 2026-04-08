CREATE TABLE users (
    id            BIGSERIAL    PRIMARY KEY,
    oidc_subject  TEXT         NOT NULL UNIQUE,
    email         TEXT         NOT NULL UNIQUE,
    name          TEXT         NOT NULL,
    avatar_url    TEXT         NOT NULL DEFAULT '',
    preferences   JSONB        NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_oidc_subject ON users (oidc_subject);
