CREATE INDEX IF NOT EXISTS idx_goal_profiles_goal_id_version_no_desc
    ON goal_profiles(goal_id, version_no DESC);
