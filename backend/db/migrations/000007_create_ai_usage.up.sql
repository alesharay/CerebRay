CREATE TABLE ai_usage_log (
    id              BIGSERIAL    PRIMARY KEY,
    user_id         BIGINT       REFERENCES users(id),
    conversation_id BIGINT       REFERENCES conversations(id),
    input_tokens    INT          NOT NULL DEFAULT 0,
    output_tokens   INT          NOT NULL DEFAULT 0,
    model           TEXT         NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_usage_user_month ON ai_usage_log (user_id, created_at);
