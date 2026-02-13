CREATE TABLE IF NOT EXISTS goal_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    template_version TEXT NOT NULL DEFAULT 'goal_brief_v1',
    version_no INT NOT NULL DEFAULT 1,
    profile_json JSONB NOT NULL DEFAULT '{}'::JSONB,
    profile_markdown TEXT NOT NULL DEFAULT '',
    completeness_score INT NOT NULL DEFAULT 0,
    open_questions JSONB NOT NULL DEFAULT '[]'::JSONB,
    confirmation_state TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT goal_profiles_template_version_chk CHECK (template_version = 'goal_brief_v1'),
    CONSTRAINT goal_profiles_confirmation_state_chk CHECK (confirmation_state IN ('pending', 'confirmed'))
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'goals_active_profile_id_fkey'
    ) THEN
        ALTER TABLE goals
            ADD CONSTRAINT goals_active_profile_id_fkey
            FOREIGN KEY (active_profile_id)
            REFERENCES goal_profiles(id)
            ON DELETE SET NULL;
    END IF;
END $$;
