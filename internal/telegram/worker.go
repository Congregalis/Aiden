package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

const (
	defaultPollFailureBackoffBase = 500 * time.Millisecond
	defaultPollFailureBackoffMax  = 8 * time.Second
)

type WorkerConfig struct {
	PollTimeoutSec int
	PollInterval   time.Duration
	AllowedUpdates []string
}

type Worker struct {
	client         Client
	store          Store
	router         Router
	sender         Sender
	logger         *slog.Logger
	pollTimeoutSec int
	pollInterval   time.Duration
	allowedUpdates []string
	metrics        *PollingMetrics
}

func NewWorker(cfg WorkerConfig, client Client, store Store, logger *slog.Logger) *Worker {
	if logger == nil {
		logger = slog.Default()
	}

	return &Worker{
		client:         client,
		store:          store,
		router:         NewRouter(),
		sender:         NewSender(client, logger),
		logger:         logger,
		pollTimeoutSec: cfg.PollTimeoutSec,
		pollInterval:   cfg.PollInterval,
		allowedUpdates: cfg.AllowedUpdates,
		metrics:        &PollingMetrics{},
	}
}

func (w *Worker) Run(ctx context.Context) error {
	me, err := w.client.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("startup getMe failed: %w", err)
	}

	w.logger.Info("telegram bot verified",
		slog.Int64("bot_id", me.ID),
		slog.String("bot_username", me.Username),
	)

	lastUpdateID, err := w.store.LoadLastUpdateID(ctx)
	if err != nil {
		return fmt.Errorf("load last update id failed: %w", err)
	}

	w.logger.Info("telegram polling worker started",
		slog.Int64("last_update_id", lastUpdateID),
		slog.Int("poll_timeout_sec", w.pollTimeoutSec),
	)

	failureStreak := 0
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("telegram polling worker stopped")
			return nil
		default:
		}

		updates, err := w.client.GetUpdates(ctx, GetUpdatesParams{
			Offset:         lastUpdateID + 1,
			TimeoutSec:     w.pollTimeoutSec,
			AllowedUpdates: w.allowedUpdates,
		})
		if err != nil {
			if ctx.Err() != nil {
				w.logger.Info("telegram polling worker stopped")
				return nil
			}

			failureStreak++
			successCount, failureCount := w.metrics.RecordFailure()
			backoff := pollingFailureBackoff(failureStreak)
			w.logger.Warn("polling_cycle_failed",
				slog.Int("failure_streak", failureStreak),
				slog.Duration("backoff", backoff),
				slog.Uint64("polling_success_count", successCount),
				slog.Uint64("polling_failure_count", failureCount),
				slog.Any("error", err),
			)

			if !sleepWithContext(ctx, backoff) {
				w.logger.Info("telegram polling worker stopped")
				return nil
			}
			continue
		}

		failureStreak = 0
		successCount, failureCount := w.metrics.RecordSuccess()
		w.logger.Info("polling_cycle_succeeded",
			slog.Int("updates_count", len(updates)),
			slog.Uint64("polling_success_count", successCount),
			slog.Uint64("polling_failure_count", failureCount),
		)

		if len(updates) == 0 {
			if !sleepWithContext(ctx, w.pollInterval) {
				w.logger.Info("telegram polling worker stopped")
				return nil
			}
			continue
		}

		for _, update := range updates {
			if err := w.handleUpdate(ctx, update); err != nil {
				w.logger.Error("handle telegram update failed",
					slog.Int64("update_id", update.UpdateID),
					slog.Any("error", err),
				)
			}

			if update.UpdateID > lastUpdateID {
				if err := w.store.SaveLastUpdateID(ctx, update.UpdateID); err != nil {
					return fmt.Errorf("persist last update id failed: %w", err)
				}
				lastUpdateID = update.UpdateID
			}
		}
	}
}

func (w *Worker) handleUpdate(ctx context.Context, update Update) error {
	message, ok := MapUpdateToIncomingMessage(update)
	if !ok {
		w.logger.Info("skip non-message update",
			slog.Int64("update_id", update.UpdateID),
		)
		return nil
	}

	chatIDMasked := MaskChatID(message.ChatID)
	w.logger.Info("telegram_update_received",
		slog.Int64("update_id", message.UpdateID),
		slog.String("chat_id_masked", chatIDMasked),
	)

	isNew, err := w.store.MarkMessageDedup(ctx, message.UpdateID, message.ChatID)
	if err != nil {
		return fmt.Errorf("message dedup failed: %w", err)
	}
	if !isNew {
		w.logger.Info("duplicate_message_skipped",
			slog.Int64("update_id", message.UpdateID),
			slog.String("chat_id_masked", chatIDMasked),
		)
		return nil
	}

	if strings.TrimSpace(message.Text) == "" {
		return w.sender.Send(ctx, OutgoingMessage{
			ChatID:           message.ChatID,
			Text:             ReplyNonText,
			ReplyToMessageID: message.MessageID,
		})
	}

	reply := w.router.ReplyFor(message)
	return w.sender.Send(ctx, OutgoingMessage{
		ChatID:           message.ChatID,
		Text:             reply,
		ReplyToMessageID: message.MessageID,
	})
}

func pollingFailureBackoff(failureStreak int) time.Duration {
	if failureStreak <= 0 {
		return defaultPollFailureBackoffBase
	}

	backoff := defaultPollFailureBackoffBase
	for i := 1; i < failureStreak; i++ {
		backoff *= 2
		if backoff >= defaultPollFailureBackoffMax {
			return defaultPollFailureBackoffMax
		}
	}

	return backoff
}
