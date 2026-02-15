package telegram

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const (
	defaultSendMaxRetries = 2
	defaultSendRetryDelay = 200 * time.Millisecond
	defaultSendRetryMax   = 2 * time.Second
)

type Sender struct {
	client     Client
	logger     *slog.Logger
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

func NewSender(client Client, logger *slog.Logger) Sender {
	if logger == nil {
		logger = slog.Default()
	}

	return Sender{
		client:     client,
		logger:     logger,
		maxRetries: defaultSendMaxRetries,
		baseDelay:  defaultSendRetryDelay,
		maxDelay:   defaultSendRetryMax,
	}
}

func (s Sender) Send(ctx context.Context, message OutgoingMessage) error {
	delay := s.baseDelay

	for attempt := 0; ; attempt++ {
		_, err := s.client.SendMessage(ctx, message)
		if err == nil {
			return nil
		}

		waitDuration, retryable := s.retryDecision(err, delay)
		if !retryable || attempt >= s.maxRetries {
			return fmt.Errorf("send message failed: %w", err)
		}

		s.logger.Warn("telegram send retry",
			slog.Int("attempt", attempt+1),
			slog.Duration("wait", waitDuration),
			slog.Any("error", err),
		)

		if !sleepWithContext(ctx, waitDuration) {
			return context.Canceled
		}

		delay *= 2
		if delay > s.maxDelay {
			delay = s.maxDelay
		}
	}
}

func (s Sender) retryDecision(err error, defaultDelay time.Duration) (time.Duration, bool) {
	retryAfter, isRateLimited := IsRateLimitError(err)
	if isRateLimited {
		return retryAfter, true
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode >= http.StatusInternalServerError {
			return defaultDelay, true
		}
		return 0, false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return defaultDelay, true
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return 0, false
	}

	return defaultDelay, true
}
