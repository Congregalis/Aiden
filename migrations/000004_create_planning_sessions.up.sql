CREATE TABLE IF NOT EXISTS planning_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    state TEXT NOT NULL DEFAULT 'idle',
    slot_completion JSONB NOT NULL DEFAULT '{}'::JSONB,
    turn_count INT NOT NULL DEFAULT 0,
    last_intent TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT planning_sessions_state_chk CHECK (state IN ('idle', 'clarifying', 'review', 'confirmed')),
    CONSTRAINT planning_sessions_turn_count_chk CHECK (turn_count >= 0)
);
