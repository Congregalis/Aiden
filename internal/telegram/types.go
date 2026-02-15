package telegram

type APIResponse[T any] struct {
	OK          bool                `json:"ok"`
	Result      T                   `json:"result"`
	Description string              `json:"description,omitempty"`
	ErrorCode   int                 `json:"error_code,omitempty"`
	Parameters  *ResponseParameters `json:"parameters,omitempty"`
}

type ResponseParameters struct {
	RetryAfter int `json:"retry_after,omitempty"`
}

type BotUser struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type GetUpdatesParams struct {
	Offset         int64
	TimeoutSec     int
	AllowedUpdates []string
}

type OutgoingMessage struct {
	ChatID           int64
	Text             string
	ReplyToMessageID int64
}

type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message,omitempty"`
}

type Message struct {
	MessageID int64  `json:"message_id"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text,omitempty"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type IncomingMessage struct {
	UpdateID  int64
	MessageID int64
	ChatID    int64
	Text      string
}
