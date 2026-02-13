CREATE INDEX IF NOT EXISTS idx_conversation_turns_session_id_created_at
    ON conversation_turns(session_id, created_at);
