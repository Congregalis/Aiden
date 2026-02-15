package telegram

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	maxFollowUpQuestionsPerTurn = 2
)

var (
	reHoursOrMinutes = regexp.MustCompile(`\d+\s*(小时|h|hr|分钟|min)`)
	reSuccessHints   = regexp.MustCompile(`(\d+\s*条|三条|四条|五条|[1-5][.、])`)
)

func UpdateSlotCompletionFromText(slotCompletion map[string]bool, text string) map[string]bool {
	normalized := NormalizeSlotCompletion(slotCompletion)
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return normalized
	}

	command := ParseCommand(trimmed)
	if command.IsCommand {
		return normalized
	}

	if detectMainGoal(trimmed) {
		normalized[SlotMainGoal] = true
	}
	if detectSuccessCriteria(trimmed) {
		normalized[SlotSuccessCriteria] = true
	}
	if detectCurrentLevel(trimmed) {
		normalized[SlotCurrentLevel] = true
	}
	if detectTimeBudget(trimmed) {
		normalized[SlotTimeBudget] = true
	}
	if detectConstraints(trimmed) {
		normalized[SlotConstraints] = true
	}
	if detectRiskFlags(trimmed) || normalized[SlotConstraints] {
		normalized[SlotRiskFlags] = true
	}

	return normalized
}

func BuildFollowUpQuestions(missingSlots []string, limit int) []string {
	if limit <= 0 {
		return nil
	}
	if limit > maxFollowUpQuestionsPerTurn {
		limit = maxFollowUpQuestionsPerTurn
	}

	questions := make([]string, 0, limit)
	for _, slot := range missingSlots {
		question := followUpQuestionBySlot(slot)
		if question == "" {
			continue
		}
		questions = append(questions, question)
		if len(questions) >= limit {
			break
		}
	}
	return questions
}

func FormatFollowUpQuestions(questions []string) string {
	if len(questions) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("我先补齐关键信息：")
	for i, question := range questions {
		builder.WriteString(fmt.Sprintf("\n%d) %s", i+1, question))
	}
	return builder.String()
}

func BuildProgressSummary(slotCompletion map[string]bool) string {
	normalized := NormalizeSlotCompletion(slotCompletion)
	missing := MissingRequiredSlots(normalized)
	filledCount := len(requiredSlotOrder) - len(missing)

	filledLabels := make([]string, 0, len(requiredSlotOrder))
	for _, slot := range requiredSlotOrder {
		if normalized[slot] {
			filledLabels = append(filledLabels, SlotLabel(slot))
		}
	}
	missingLabels := make([]string, 0, len(missing))
	for _, slot := range missing {
		missingLabels = append(missingLabels, SlotLabel(slot))
	}

	filledText := "无"
	if len(filledLabels) > 0 {
		filledText = strings.Join(filledLabels, "、")
	}
	missingText := "无"
	if len(missingLabels) > 0 {
		missingText = strings.Join(missingLabels, "、")
	}

	return fmt.Sprintf("【当前摘要】已补齐 %d/%d 项：%s；待补齐：%s。\n当前版本你是否满意，还是继续优化？",
		filledCount,
		len(requiredSlotOrder),
		filledText,
		missingText,
	)
}

func followUpQuestionBySlot(slot string) string {
	switch slot {
	case SlotMainGoal:
		return "你希望在什么时间前达成什么主目标？"
	case SlotSuccessCriteria:
		return "请给我 3-5 条可验收的成功标准（尽量量化）。"
	case SlotCurrentLevel:
		return "你当前水平如何（零基础/入门/有项目经验）？"
	case SlotTimeBudget:
		return "你每周可投入多少小时，或有哪些固定学习时段？"
	case SlotConstraints:
		return "有哪些约束会影响执行（如加班、设备、可用时段）？"
	case SlotRiskFlags:
		return "你担心哪些风险会影响坚持（如出差、拖延、突发事务）？"
	default:
		return ""
	}
}

func detectMainGoal(text string) bool {
	if len([]rune(strings.TrimSpace(text))) < 6 {
		return false
	}
	return containsAny(text, []string{"目标", "我想", "希望", "计划", "完成", "学会", "掌握", "通过", "提升"})
}

func detectSuccessCriteria(text string) bool {
	if reSuccessHints.MatchString(text) {
		return true
	}
	return containsAny(text, []string{"成功标准", "验收", "里程碑", "达到", "完成", "通过"})
}

func detectCurrentLevel(text string) bool {
	return containsAny(text, []string{
		"零基础", "新手", "入门", "初级", "中级", "高级",
		"不会", "有经验", "做过项目", "基础薄弱",
	})
}

func detectTimeBudget(text string) bool {
	if reHoursOrMinutes.MatchString(text) {
		return true
	}
	return containsAny(text, []string{
		"每周", "每天", "工作日", "周末", "晚上", "早上", "午休", "通勤",
	})
}

func detectConstraints(text string) bool {
	return containsAny(text, []string{
		"只能", "没时间", "限制", "约束", "加班", "带娃", "出差", "设备", "网络", "时间不固定",
	})
}

func detectRiskFlags(text string) bool {
	return containsAny(text, []string{
		"风险", "担心", "拖延", "中断", "坚持不下去", "突发", "不稳定", "焦虑", "压力",
	})
}
