CREATE TABLE IF NOT EXISTS message_dedup (
    update_id BIGINT PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
