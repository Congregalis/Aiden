CREATE TABLE IF NOT EXISTS conversation_turns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES planning_sessions(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    intent TEXT NOT NULL DEFAULT '',
    intent_confidence NUMERIC(5,4),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT conversation_turns_role_chk CHECK (role IN ('user', 'assistant', 'system')),
    CONSTRAINT conversation_turns_intent_confidence_chk CHECK (
        intent_confidence IS NULL OR
        (intent_confidence >= 0 AND intent_confidence <= 1)
    )
);
