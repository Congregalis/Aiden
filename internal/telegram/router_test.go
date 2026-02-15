package telegram

import "testing"

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		isCommand bool
	}{
		{name: "start command", input: "/start", expected: "start", isCommand: true},
		{name: "command with bot mention", input: "/goal@AidenBot", expected: "goal", isCommand: true},
		{name: "natural language", input: "我想学 Go", expected: "", isCommand: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed := ParseCommand(tc.input)
			if parsed.IsCommand != tc.isCommand {
				t.Fatalf("ParseCommand(%q) isCommand=%v, want %v", tc.input, parsed.IsCommand, tc.isCommand)
			}
			if parsed.Name != tc.expected {
				t.Fatalf("ParseCommand(%q) name=%q, want %q", tc.input, parsed.Name, tc.expected)
			}
		})
	}
}

func TestRouterReplyFor(t *testing.T) {
	router := NewRouter()

	tests := []struct {
		name     string
		message  IncomingMessage
		expected string
	}{
		{name: "help", message: IncomingMessage{Text: "/help"}, expected: ReplyHelp},
		{name: "unknown command", message: IncomingMessage{Text: "/plan"}, expected: ReplyUnknownCommand},
		{name: "natural language", message: IncomingMessage{Text: "我想在两个月内学会Go"}, expected: ReplyNaturalMessage},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := router.ReplyFor(tc.message)
			if actual != tc.expected {
				t.Fatalf("ReplyFor()=%q, want %q", actual, tc.expected)
			}
		})
	}
}

func TestRouterReplyForStart(t *testing.T) {
	router := NewRouter()

	if got := router.ReplyForStart(true); got != ReplyStart {
		t.Fatalf("ReplyForStart(true)=%q, want %q", got, ReplyStart)
	}
	if got := router.ReplyForStart(false); got != ReplyStartBack {
		t.Fatalf("ReplyForStart(false)=%q, want %q", got, ReplyStartBack)
	}
}
