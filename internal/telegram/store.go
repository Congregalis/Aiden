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
}

type SQLStore struct {
	db *sql.DB
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
