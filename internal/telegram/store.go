package telegram

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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
