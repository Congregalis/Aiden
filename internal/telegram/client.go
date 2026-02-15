package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultTelegramBaseURL = "https://api.telegram.org"

type Client interface {
	GetMe(context.Context) (BotUser, error)
	GetUpdates(context.Context, GetUpdatesParams) ([]Update, error)
	SendMessage(context.Context, OutgoingMessage) (Message, error)
}

type HTTPClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewHTTPClient(token string, httpClient *http.Client) *HTTPClient {
	return NewHTTPClientWithBaseURL(token, defaultTelegramBaseURL, httpClient)
}

func NewHTTPClientWithBaseURL(token, baseURL string, httpClient *http.Client) *HTTPClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 70 * time.Second}
	}

	cleanBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if cleanBaseURL == "" {
		cleanBaseURL = defaultTelegramBaseURL
	}

	return &HTTPClient{
		baseURL:    cleanBaseURL,
		token:      strings.TrimSpace(token),
		httpClient: httpClient,
	}
}

func (c *HTTPClient) GetMe(ctx context.Context) (BotUser, error) {
	body, statusCode, err := c.postJSON(ctx, "getMe", map[string]any{})
	if err != nil {
		return BotUser{}, err
	}

	result, err := decodeResult[BotUser](statusCode, body)
	if err != nil {
		return BotUser{}, fmt.Errorf("telegram getMe: %w", err)
	}

	return result, nil
}

func (c *HTTPClient) GetUpdates(ctx context.Context, params GetUpdatesParams) ([]Update, error) {
	request := map[string]any{
		"timeout": params.TimeoutSec,
	}
	if params.Offset > 0 {
		request["offset"] = params.Offset
	}
	if len(params.AllowedUpdates) > 0 {
		request["allowed_updates"] = params.AllowedUpdates
	}

	body, statusCode, err := c.postJSON(ctx, "getUpdates", request)
	if err != nil {
		return nil, err
	}

	result, err := decodeResult[[]Update](statusCode, body)
	if err != nil {
		return nil, fmt.Errorf("telegram getUpdates: %w", err)
	}

	return result, nil
}

func (c *HTTPClient) SendMessage(ctx context.Context, message OutgoingMessage) (Message, error) {
	request := map[string]any{
		"chat_id": message.ChatID,
		"text":    message.Text,
	}
	if message.ReplyToMessageID > 0 {
		request["reply_to_message_id"] = message.ReplyToMessageID
	}

	body, statusCode, err := c.postJSON(ctx, "sendMessage", request)
	if err != nil {
		return Message{}, err
	}

	result, err := decodeResult[Message](statusCode, body)
	if err != nil {
		return Message{}, fmt.Errorf("telegram sendMessage: %w", err)
	}

	return result, nil
}

func (c *HTTPClient) postJSON(ctx context.Context, method string, payload map[string]any) ([]byte, int, error) {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal telegram %s request: %w", method, err)
	}

	url := fmt.Sprintf("%s/bot%s/%s", c.baseURL, c.token, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, 0, fmt.Errorf("build telegram %s request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("telegram %s request: %w", method, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read telegram %s response: %w", method, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, fmt.Errorf("telegram %s: %w", method, parseAPIError(resp.StatusCode, body))
	}

	return body, resp.StatusCode, nil
}

func decodeResult[T any](statusCode int, body []byte) (T, error) {
	var zero T
	var envelope APIResponse[T]
	if err := json.Unmarshal(body, &envelope); err != nil {
		return zero, fmt.Errorf("decode telegram response: %w", err)
	}

	if !envelope.OK {
		return zero, &APIError{
			StatusCode:  statusCode,
			ErrorCode:   envelope.ErrorCode,
			Description: envelope.Description,
			RetryAfter:  retryAfterFromParameters(envelope.Parameters),
		}
	}

	return envelope.Result, nil
}

func retryAfterFromParameters(params *ResponseParameters) time.Duration {
	if params == nil || params.RetryAfter <= 0 {
		return 0
	}
	return time.Duration(params.RetryAfter) * time.Second
}

func parseAPIError(statusCode int, body []byte) error {
	var envelope APIResponse[json.RawMessage]
	if err := json.Unmarshal(body, &envelope); err != nil {
		return &APIError{
			StatusCode:  statusCode,
			Description: strings.TrimSpace(string(body)),
		}
	}

	return &APIError{
		StatusCode:  statusCode,
		ErrorCode:   envelope.ErrorCode,
		Description: envelope.Description,
		RetryAfter:  retryAfterFromParameters(envelope.Parameters),
	}
}

type APIError struct {
	StatusCode  int
	ErrorCode   int
	Description string
	RetryAfter  time.Duration
}

func (e *APIError) Error() string {
	if e == nil {
		return "telegram api error"
	}
	if e.Description == "" {
		return fmt.Sprintf("telegram api error (status=%d, code=%d)", e.StatusCode, e.ErrorCode)
	}
	return fmt.Sprintf("telegram api error (status=%d, code=%d): %s", e.StatusCode, e.ErrorCode, e.Description)
}

func IsRateLimitError(err error) (time.Duration, bool) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return 0, false
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		return 0, false
	}
	if apiErr.RetryAfter <= 0 {
		return time.Second, true
	}
	return apiErr.RetryAfter, true
}
