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
	defaultSessionTimeout         = 24 * time.Hour
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
	intentRouter   IntentRouter
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
		intentRouter:   NewIntentRouter(),
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

	user, isNewUser, err := w.store.FindOrCreateUserByChatID(ctx, message.ChatID)
	if err != nil {
		return fmt.Errorf("find or create user by chat id: %w", err)
	}

	command := ParseCommand(message.Text)
	reply := ReplyNaturalMessage
	if command.IsCommand {
		switch command.Name {
		case "start":
			reply = w.router.ReplyForStart(isNewUser)
		case "goal":
			goal, err := w.ensureActiveGoal(ctx, user, message.ChatID)
			if err != nil {
				return err
			}
			reply, err = w.handleClarifyRound(ctx, message, goal)
			if err != nil {
				return err
			}
		case "help":
			reply = ReplyHelp
		default:
			reply = ReplyUnknownCommand
		}
	} else {
		goal, err := w.ensureActiveGoal(ctx, user, message.ChatID)
		if err != nil {
			return err
		}
		reply, err = w.handleClarifyRound(ctx, message, goal)
		if err != nil {
			return err
		}
	}

	return w.sender.Send(ctx, OutgoingMessage{
		ChatID:           message.ChatID,
		Text:             reply,
		ReplyToMessageID: message.MessageID,
	})
}

func (w *Worker) ensureActiveGoal(ctx context.Context, user User, chatID int64) (Goal, error) {
	goal, found, err := w.store.GetActiveGoalByUserID(ctx, user.ID)
	if err != nil {
		return Goal{}, fmt.Errorf("get active goal by user id: %w", err)
	}
	if found {
		return goal, nil
	}

	createdGoal, err := w.store.CreateGoalDraft(ctx, user.ID)
	if err != nil {
		return Goal{}, fmt.Errorf("create goal draft: %w", err)
	}

	w.logger.Info("goal_started",
		slog.String("goal_id", createdGoal.ID),
		slog.String("user_id", createdGoal.UserID),
		slog.String("goal_status", createdGoal.Status),
		slog.String("chat_id_masked", MaskChatID(chatID)),
	)

	return createdGoal, nil
}

func (w *Worker) handleClarifyRound(ctx context.Context, message IncomingMessage, goal Goal) (string, error) {
	session, _, err := w.store.GetOrCreatePlanningSession(ctx, goal.ID)
	if err != nil {
		return "", fmt.Errorf("get or create planning session: %w", err)
	}

	timeoutNotice := ""
	if shouldResetSessionForTimeout(session.UpdatedAt) && !session.State.IsFinal() {
		timeoutNotice = ReplySessionTimeout
		session.State = StateClarifying
		session.LastIntent = ""
		session.TurnCount = 0
		if err := w.store.UpdatePlanningSession(ctx, session); err != nil {
			return "", fmt.Errorf("reset planning session after timeout: %w", err)
		}
	}

	intent := w.intentRouter.Route(message.Text, session.State)

	turnCount, err := w.store.IncrementPlanningSessionTurn(ctx, session.ID)
	if err != nil {
		return "", fmt.Errorf("increment planning session turn: %w", err)
	}
	session.TurnCount = turnCount

	if err := w.store.SaveConversationTurn(ctx, ConversationTurn{
		SessionID:        session.ID,
		Role:             ConversationRoleUser,
		Content:          message.Text,
		Intent:           intent.Intent,
		IntentConfidence: &intent.Confidence,
	}); err != nil {
		return "", fmt.Errorf("save user conversation turn: %w", err)
	}

	reply, updatedSession := w.buildClarifyReply(session, message.Text, intent)
	if timeoutNotice != "" {
		reply = timeoutNotice + "\n\n" + reply
	}
	if updatedSession.TurnCount > 0 &&
		updatedSession.TurnCount%3 == 0 &&
		!updatedSession.State.IsFinal() &&
		!strings.Contains(reply, "【当前摘要】") {
		reply = reply + "\n\n" + BuildProgressSummary(updatedSession.SlotCompletion)
	}

	if err := w.store.UpdatePlanningSession(ctx, updatedSession); err != nil {
		return "", fmt.Errorf("update planning session: %w", err)
	}

	if err := w.store.SaveConversationTurn(ctx, ConversationTurn{
		SessionID: updatedSession.ID,
		Role:      ConversationRoleAssistant,
		Content:   reply,
		Intent:    intent.Intent,
	}); err != nil {
		return "", fmt.Errorf("save assistant conversation turn: %w", err)
	}

	return reply, nil
}

func (w *Worker) buildClarifyReply(session PlanningSession, text string, intent IntentResult) (string, PlanningSession) {
	updated := session
	updated.State = ParsePlanningState(string(updated.State))
	if updated.State == StateIdle {
		updated.State = StateClarifying
	}
	updated.SlotCompletion = NormalizeSlotCompletion(updated.SlotCompletion)
	updated.LastIntent = intent.Intent

	command := ParseCommand(text)
	isGoalCommand := command.IsCommand && command.Name == "goal"

	if isGoalCommand {
		updated.State = StateClarifying
		return ReplyGoal, updated
	}

	shouldExtractSlots := intent.Intent == IntentClarifyGoal
	if shouldExtractSlots {
		updated.SlotCompletion = UpdateSlotCompletionFromText(updated.SlotCompletion, text)
	}

	switch updated.State {
	case StateReview:
		if intent.Intent == IntentConfirmPlan {
			updated.State = StateConfirmed
			return ReplyPlanConfirmed, updated
		}
		if intent.Intent == IntentClarifyGoal {
			updated.State = StateClarifying
			questions := BuildFollowUpQuestions(MissingRequiredSlots(updated.SlotCompletion), 2)
			if len(questions) == 0 {
				return "已收到修改意见，我已切回 clarifying。请补充你希望调整的重点。", updated
			}
			return "已收到修改意见，我已切回 clarifying。\n" + FormatFollowUpQuestions(questions), updated
		}
		return ReplyReviewFallback, updated

	case StateConfirmed:
		if intent.Intent == IntentClarifyGoal {
			updated.State = StateClarifying
			questions := BuildFollowUpQuestions(MissingRequiredSlots(updated.SlotCompletion), 1)
			if len(questions) == 0 {
				return "我已重新打开澄清会话，请告诉我你想调整的目标内容。", updated
			}
			return "我已重新打开澄清会话。\n" + FormatFollowUpQuestions(questions), updated
		}
		return ReplyPlanConfirmed, updated
	}

	if intent.Intent == IntentFallbackUnknown {
		return ReplyFallbackGuidance, updated
	}

	if IsRequiredSlotsComplete(updated.SlotCompletion) {
		updated.State = StateReview
		return ReplyReviewReady + "\n\n" + BuildProgressSummary(updated.SlotCompletion), updated
	}

	questions := BuildFollowUpQuestions(MissingRequiredSlots(updated.SlotCompletion), 2)
	if len(questions) == 0 {
		return ReplyNaturalMessage, updated
	}

	return FormatFollowUpQuestions(questions), updated
}

func shouldResetSessionForTimeout(lastUpdatedAt time.Time) bool {
	if lastUpdatedAt.IsZero() {
		return false
	}
	return time.Since(lastUpdatedAt) >= defaultSessionTimeout
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
