package telegram

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

const (
	runtimeOffsetKey       = "telegram_last_update_id"
	legacyRuntimeOffsetKey = "last_update_id"
)

type Store interface {
	LoadLastUpdateID(context.Context) (int64, error)
	SaveLastUpdateID(context.Context, int64) error
	MarkMessageDedup(context.Context, int64, int64) (bool, error)
	FindOrCreateUserByChatID(context.Context, int64) (User, bool, error)
	GetActiveGoalByUserID(context.Context, string) (Goal, bool, error)
	CreateGoalDraft(context.Context, string) (Goal, error)
	GetOrCreatePlanningSession(context.Context, string) (PlanningSession, bool, error)
	IncrementPlanningSessionTurn(context.Context, string) (int, error)
	UpdatePlanningSession(context.Context, PlanningSession) error
	SaveConversationTurn(context.Context, ConversationTurn) error
}

type SQLStore struct {
	db *sql.DB
}

type User struct {
	ID             string
	TelegramChatID int64
	Language       string
	Timezone       string
}

type Goal struct {
	ID     string
	UserID string
	Title  string
	Status string
}

type PlanningSession struct {
	ID             string
	GoalID         string
	State          PlanningState
	SlotCompletion map[string]bool
	TurnCount      int
	LastIntent     string
	UpdatedAt      time.Time
}

type ConversationTurn struct {
	SessionID        string
	Role             string
	Content          string
	Intent           string
	IntentConfidence *float64
}

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

func (s *SQLStore) LoadLastUpdateID(ctx context.Context) (int64, error) {
	value, err := s.loadRuntimeStateValue(ctx, runtimeOffsetKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			legacyValue, legacyErr := s.loadRuntimeStateValue(ctx, legacyRuntimeOffsetKey)
			if errors.Is(legacyErr, sql.ErrNoRows) {
				return 0, nil
			}
			if legacyErr != nil {
				return 0, fmt.Errorf("load runtime state %s: %w", legacyRuntimeOffsetKey, legacyErr)
			}
			return parseRuntimeOffset(legacyValue)
		}
		return 0, fmt.Errorf("load runtime state %s: %w", runtimeOffsetKey, err)
	}

	return parseRuntimeOffset(value)
}

func (s *SQLStore) SaveLastUpdateID(ctx context.Context, lastUpdateID int64) error {
	value := strconv.FormatInt(lastUpdateID, 10)

	if err := s.upsertRuntimeStateValue(ctx, runtimeOffsetKey, value); err != nil {
		return fmt.Errorf("save runtime state %s: %w", runtimeOffsetKey, err)
	}

	if err := s.upsertRuntimeStateValue(ctx, legacyRuntimeOffsetKey, value); err != nil {
		return fmt.Errorf("save runtime state %s: %w", legacyRuntimeOffsetKey, err)
	}

	return nil
}

func (s *SQLStore) MarkMessageDedup(ctx context.Context, updateID, chatID int64) (bool, error) {
	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO message_dedup(update_id, chat_id, received_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (update_id) DO NOTHING`,
		updateID,
		chatID,
	)
	if err != nil {
		return false, fmt.Errorf("insert message_dedup: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read dedup rows affected: %w", err)
	}

	return rowsAffected == 1, nil
}

func (s *SQLStore) FindOrCreateUserByChatID(ctx context.Context, chatID int64) (User, bool, error) {
	createdUser, err := s.insertUserIfNotExists(ctx, chatID)
	if err == nil {
		return createdUser, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return User{}, false, fmt.Errorf("insert user by chat id %d: %w", chatID, err)
	}

	existingUser, err := s.findUserByChatID(ctx, chatID)
	if err != nil {
		return User{}, false, fmt.Errorf("find user by chat id %d: %w", chatID, err)
	}

	return existingUser, false, nil
}

func (s *SQLStore) GetActiveGoalByUserID(ctx context.Context, userID string) (Goal, bool, error) {
	var goal Goal
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, title, status
		 FROM goals
		 WHERE user_id = $1
		   AND status IN ('active', 'draft')
		 ORDER BY
		   CASE status
		     WHEN 'active' THEN 0
		     WHEN 'draft' THEN 1
		     ELSE 2
		   END,
		   updated_at DESC,
		   created_at DESC
		 LIMIT 1`,
		userID,
	).Scan(&goal.ID, &goal.UserID, &goal.Title, &goal.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return Goal{}, false, nil
	}
	if err != nil {
		return Goal{}, false, fmt.Errorf("query active goal by user id %s: %w", userID, err)
	}

	return goal, true, nil
}

func (s *SQLStore) CreateGoalDraft(ctx context.Context, userID string) (Goal, error) {
	var goal Goal
	err := s.db.QueryRowContext(
		ctx,
		`INSERT INTO goals(user_id, title, status, created_at, updated_at)
		 VALUES ($1, '', 'draft', NOW(), NOW())
		 RETURNING id, user_id, title, status`,
		userID,
	).Scan(&goal.ID, &goal.UserID, &goal.Title, &goal.Status)
	if err != nil {
		return Goal{}, fmt.Errorf("insert goal draft for user id %s: %w", userID, err)
	}

	return goal, nil
}

func (s *SQLStore) GetOrCreatePlanningSession(ctx context.Context, goalID string) (PlanningSession, bool, error) {
	created, err := s.insertPlanningSessionIfNotExists(ctx, goalID)
	if err == nil {
		return created, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return PlanningSession{}, false, fmt.Errorf("insert planning session for goal id %s: %w", goalID, err)
	}

	existing, err := s.findPlanningSessionByGoalID(ctx, goalID)
	if err != nil {
		return PlanningSession{}, false, fmt.Errorf("find planning session by goal id %s: %w", goalID, err)
	}

	return existing, false, nil
}

func (s *SQLStore) IncrementPlanningSessionTurn(ctx context.Context, sessionID string) (int, error) {
	var turnCount int
	err := s.db.QueryRowContext(
		ctx,
		`UPDATE planning_sessions
		 SET turn_count = turn_count + 1,
		     updated_at = NOW()
		 WHERE id = $1
		 RETURNING turn_count`,
		sessionID,
	).Scan(&turnCount)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("planning session %s not found", sessionID)
	}
	if err != nil {
		return 0, fmt.Errorf("increment planning session turn for session id %s: %w", sessionID, err)
	}

	return turnCount, nil
}

func (s *SQLStore) UpdatePlanningSession(ctx context.Context, session PlanningSession) error {
	slotCompletionJSON, err := json.Marshal(NormalizeSlotCompletion(session.SlotCompletion))
	if err != nil {
		return fmt.Errorf("marshal slot completion: %w", err)
	}

	result, err := s.db.ExecContext(
		ctx,
		`UPDATE planning_sessions
		 SET state = $2,
		     slot_completion = $3::jsonb,
		     turn_count = $4,
		     last_intent = $5,
		     updated_at = NOW()
		 WHERE id = $1`,
		session.ID,
		string(session.State),
		slotCompletionJSON,
		session.TurnCount,
		session.LastIntent,
	)
	if err != nil {
		return fmt.Errorf("update planning session %s: %w", session.ID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read update planning session rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("planning session %s not found", session.ID)
	}

	return nil
}

func (s *SQLStore) SaveConversationTurn(ctx context.Context, turn ConversationTurn) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO conversation_turns(
		    session_id,
		    role,
		    content,
		    intent,
		    intent_confidence,
		    created_at
		 )
		 VALUES ($1, $2, $3, $4, $5, NOW())`,
		turn.SessionID,
		turn.Role,
		turn.Content,
		turn.Intent,
		turn.IntentConfidence,
	)
	if err != nil {
		return fmt.Errorf("insert conversation turn: %w", err)
	}

	return nil
}

func (s *SQLStore) loadRuntimeStateValue(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(
		ctx,
		`SELECT value FROM bot_runtime_states WHERE key = $1`,
		key,
	).Scan(&value)
	if err != nil {
		return "", err
	}

	return value, nil
}

func (s *SQLStore) upsertRuntimeStateValue(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO bot_runtime_states(key, value, updated_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (key) DO UPDATE
		 SET value = EXCLUDED.value,
		     updated_at = NOW()`,
		key,
		value,
	)
	return err
}

func parseRuntimeOffset(value string) (int64, error) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse runtime offset %q: %w", value, err)
	}
	if parsed < 0 {
		return 0, fmt.Errorf("runtime offset must be >= 0")
	}
	return parsed, nil
}

func (s *SQLStore) insertUserIfNotExists(ctx context.Context, chatID int64) (User, error) {
	var user User
	err := s.db.QueryRowContext(
		ctx,
		`INSERT INTO users(telegram_chat_id, language, timezone, created_at, updated_at)
		 VALUES ($1, 'zh-CN', 'Asia/Shanghai', NOW(), NOW())
		 ON CONFLICT (telegram_chat_id) DO NOTHING
		 RETURNING id, telegram_chat_id, language, timezone`,
		chatID,
	).Scan(&user.ID, &user.TelegramChatID, &user.Language, &user.Timezone)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *SQLStore) findUserByChatID(ctx context.Context, chatID int64) (User, error) {
	var user User
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, telegram_chat_id, language, timezone
		 FROM users
		 WHERE telegram_chat_id = $1`,
		chatID,
	).Scan(&user.ID, &user.TelegramChatID, &user.Language, &user.Timezone)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *SQLStore) insertPlanningSessionIfNotExists(ctx context.Context, goalID string) (PlanningSession, error) {
	var (
		session            PlanningSession
		stateRaw           string
		slotCompletionJSON []byte
	)

	err := s.db.QueryRowContext(
		ctx,
		`INSERT INTO planning_sessions(
		    goal_id,
		    state,
		    slot_completion,
		    turn_count,
		    last_intent,
		    updated_at
		 )
		 VALUES ($1, 'idle', '{}'::jsonb, 0, '', NOW())
		 ON CONFLICT (goal_id) DO NOTHING
		 RETURNING id, goal_id, state, slot_completion, turn_count, last_intent, updated_at`,
		goalID,
	).Scan(
		&session.ID,
		&session.GoalID,
		&stateRaw,
		&slotCompletionJSON,
		&session.TurnCount,
		&session.LastIntent,
		&session.UpdatedAt,
	)
	if err != nil {
		return PlanningSession{}, err
	}

	slotCompletion, err := parseSlotCompletionJSON(slotCompletionJSON)
	if err != nil {
		return PlanningSession{}, fmt.Errorf("parse slot completion from created planning session: %w", err)
	}

	session.State = ParsePlanningState(stateRaw)
	session.SlotCompletion = slotCompletion
	return session, nil
}

func (s *SQLStore) findPlanningSessionByGoalID(ctx context.Context, goalID string) (PlanningSession, error) {
	var (
		session            PlanningSession
		stateRaw           string
		slotCompletionJSON []byte
	)

	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, goal_id, state, slot_completion, turn_count, last_intent, updated_at
		 FROM planning_sessions
		 WHERE goal_id = $1`,
		goalID,
	).Scan(
		&session.ID,
		&session.GoalID,
		&stateRaw,
		&slotCompletionJSON,
		&session.TurnCount,
		&session.LastIntent,
		&session.UpdatedAt,
	)
	if err != nil {
		return PlanningSession{}, err
	}

	slotCompletion, err := parseSlotCompletionJSON(slotCompletionJSON)
	if err != nil {
		return PlanningSession{}, fmt.Errorf("parse slot completion from planning session: %w", err)
	}

	session.State = ParsePlanningState(stateRaw)
	session.SlotCompletion = slotCompletion
	return session, nil
}

func parseSlotCompletionJSON(raw []byte) (map[string]bool, error) {
	if len(raw) == 0 {
		return DefaultSlotCompletion(), nil
	}

	decoded := make(map[string]bool)
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}

	return NormalizeSlotCompletion(decoded), nil
}
