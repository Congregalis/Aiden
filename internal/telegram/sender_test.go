package telegram

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestSenderRetriesOnRateLimit(t *testing.T) {
	client := &sendStubClient{
		errors: []error{
			&APIError{StatusCode: http.StatusTooManyRequests, RetryAfter: 5 * time.Millisecond},
			nil,
		},
	}

	sender := NewSender(client, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := sender.Send(ctx, OutgoingMessage{ChatID: 1, Text: "hello"})
	if err != nil {
		t.Fatalf("Send() returned error: %v", err)
	}
	if client.sendAttempts != 2 {
		t.Fatalf("send attempts=%d, want 2", client.sendAttempts)
	}
}

func TestSenderStopsOnClientError(t *testing.T) {
	client := &sendStubClient{
		errors: []error{
			&APIError{StatusCode: http.StatusBadRequest, Description: "bad request"},
		},
	}

	sender := NewSender(client, slog.New(slog.NewTextHandler(io.Discard, nil)))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := sender.Send(ctx, OutgoingMessage{ChatID: 1, Text: "hello"})
	if err == nil {
		t.Fatalf("Send() expected error")
	}
	if client.sendAttempts != 1 {
		t.Fatalf("send attempts=%d, want 1", client.sendAttempts)
	}
}

type sendStubClient struct {
	errors       []error
	sendAttempts int
}

func (c *sendStubClient) GetMe(context.Context) (BotUser, error) {
	return BotUser{}, nil
}

func (c *sendStubClient) GetUpdates(context.Context, GetUpdatesParams) ([]Update, error) {
	return nil, nil
}

func (c *sendStubClient) SendMessage(context.Context, OutgoingMessage) (Message, error) {
	idx := c.sendAttempts
	c.sendAttempts++
	if idx >= len(c.errors) {
		return Message{}, nil
	}
	if c.errors[idx] != nil {
		return Message{}, c.errors[idx]
	}
	return Message{}, nil
}
