ALTER TABLE IF EXISTS goals
    DROP CONSTRAINT IF EXISTS goals_active_profile_id_fkey;

DROP TABLE IF EXISTS goal_profiles;
