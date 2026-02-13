-- Optional seed data for local development.
-- Usage: psql "$DB_DSN" -v ON_ERROR_STOP=1 -f migrations/seed_m1.sql

INSERT INTO users (id, telegram_chat_id, timezone, language)
VALUES
    ('00000000-0000-0000-0000-000000000001', 10000001, 'Asia/Shanghai', 'zh-CN')
ON CONFLICT (telegram_chat_id) DO NOTHING;

INSERT INTO goals (id, user_id, title, status)
VALUES
    ('00000000-0000-0000-0000-000000000101', '00000000-0000-0000-0000-000000000001', '8周完成Go基础', 'draft')
ON CONFLICT (id) DO NOTHING;

INSERT INTO planning_sessions (id, goal_id, state, slot_completion, turn_count, last_intent)
VALUES
    (
        '00000000-0000-0000-0000-000000000201',
        '00000000-0000-0000-0000-000000000101',
        'clarifying',
        '{"main_goal": false, "success_criteria": false}'::JSONB,
        0,
        ''
    )
ON CONFLICT (id) DO NOTHING;

INSERT INTO bot_runtime_states (key, value)
VALUES ('last_update_id', '0')
ON CONFLICT (key) DO UPDATE
SET
    value = EXCLUDED.value,
    updated_at = NOW();
