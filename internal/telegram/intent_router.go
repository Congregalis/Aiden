package telegram

import "strings"

type IntentRouter struct{}

func NewIntentRouter() IntentRouter {
	return IntentRouter{}
}

func (r IntentRouter) Route(text string, state PlanningState) IntentResult {
	command := ParseCommand(text)
	if command.IsCommand {
		switch command.Name {
		case "goal":
			return IntentResult{Intent: IntentClarifyGoal, Confidence: 1}
		case "confirm":
			return IntentResult{Intent: IntentConfirmPlan, Confidence: 1}
		default:
			return IntentResult{Intent: IntentFallbackUnknown, Confidence: 0.7}
		}
	}

	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return IntentResult{Intent: IntentFallbackUnknown, Confidence: 0.1}
	}

	if containsAny(trimmed, confirmSignals) {
		return IntentResult{Intent: IntentConfirmPlan, Confidence: 0.92}
	}

	if containsAny(trimmed, clarifySignals) || len([]rune(trimmed)) >= 8 {
		return IntentResult{Intent: IntentClarifyGoal, Confidence: 0.78}
	}

	if state == StateReview {
		return IntentResult{Intent: IntentFallbackUnknown, Confidence: 0.45}
	}

	return IntentResult{Intent: IntentFallbackUnknown, Confidence: 0.35}
}

var confirmSignals = []string{
	"确认",
	"同意",
	"就这样",
	"没问题",
	"可以开始",
	"开始执行",
	"ok",
	"yes",
	"confirm",
}

var clarifySignals = []string{
	"目标",
	"我想",
	"计划",
	"每周",
	"小时",
	"分钟",
	"约束",
	"限制",
	"水平",
	"标准",
	"修改",
	"调整",
	"优化",
	"补充",
}

func containsAny(text string, keywords []string) bool {
	lower := strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(lower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}
