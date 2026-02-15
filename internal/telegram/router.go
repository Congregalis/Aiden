package telegram

import "strings"

const (
	ReplyStart = "你好，我是 Aiden。你可以通过 /goal 开始目标澄清，我会一步步帮你把目标讲清楚。"
	ReplyGoal  = "好的，我们开始目标澄清。请先告诉我：你希望在什么时间前达成什么目标？"
	ReplyHelp  = "当前可用命令：/start、/goal、/help。你也可以直接用自然语言告诉我你的目标。"

	ReplyNonText        = "我目前只能处理文本消息，请发送文字内容。"
	ReplyUnknownCommand = "这个命令会在后续里程碑开放。当前可用：/start、/goal、/help。"
	ReplyNaturalMessage = "收到，我已进入自然语言澄清入口。你可以继续描述目标细节，或发送 /goal 切换到命令入口。"
)

type Command struct {
	Name      string
	IsCommand bool
}

func ParseCommand(text string) Command {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" || !strings.HasPrefix(trimmed, "/") {
		return Command{}
	}

	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return Command{}
	}

	name := strings.TrimPrefix(fields[0], "/")
	if idx := strings.Index(name, "@"); idx >= 0 {
		name = name[:idx]
	}

	name = strings.ToLower(strings.TrimSpace(name))
	return Command{Name: name, IsCommand: true}
}

type Router struct{}

func NewRouter() Router {
	return Router{}
}

func (r Router) ReplyFor(message IncomingMessage) string {
	command := ParseCommand(message.Text)
	if !command.IsCommand {
		return ReplyNaturalMessage
	}

	switch command.Name {
	case "start":
		return ReplyStart
	case "goal":
		return ReplyGoal
	case "help":
		return ReplyHelp
	default:
		return ReplyUnknownCommand
	}
}
