package telegram

import "strings"

const (
	ReplyStart     = "欢迎来到 Aiden！我已完成初始化（默认语言 zh-CN，默认时区 Asia/Shanghai）。发送 /goal 开始目标澄清。"
	ReplyStartBack = "欢迎回来！发送 /goal 继续目标澄清，或发送 /help 查看可用命令。"
	ReplyGoal      = "好的，我们开始目标澄清。请先告诉我：你希望在什么时间前达成什么目标？"
	ReplyHelp      = "当前可用命令：/start、/goal、/help。你也可以直接用自然语言告诉我你的目标。"

	ReplyNonText        = "我目前只能处理文本消息，请发送文字内容。"
	ReplyUnknownCommand = "这个命令会在后续里程碑开放。当前可用：/start、/goal、/help。"
	ReplyNaturalMessage = "收到，我已进入自然语言澄清入口。你可以继续描述目标细节，或发送 /goal 切换到命令入口。"
	ReplyReviewReady    = "关键信息已补齐，我已切换到 review。回复“确认”即可完成澄清；如需修改，请直接告诉我你要调整的内容。"
	ReplyPlanConfirmed  = "已确认，当前会话状态更新为 confirmed。接下来我会按这个目标继续推进。"

	ReplyFallbackGuidance = "我这条没有完全理解。你可以直接补充：主目标、成功标准、当前水平、时间预算或约束；我会保留当前上下文继续澄清。"
	ReplyReviewFallback   = "如果你认可当前版本，请回复“确认”；如果要改动，直接说“修改 + 你的新要求”。我会保留上下文。"
	ReplySessionTimeout   = "距离上次澄清已超过 24 小时，我先帮你恢复到澄清状态，我们继续补齐信息。"
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

func (r Router) ReplyForStart(isNewUser bool) string {
	if isNewUser {
		return ReplyStart
	}
	return ReplyStartBack
}
