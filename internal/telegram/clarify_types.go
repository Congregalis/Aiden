package telegram

type PlanningState string

const (
	StateIdle       PlanningState = "idle"
	StateClarifying PlanningState = "clarifying"
	StateReview     PlanningState = "review"
	StateConfirmed  PlanningState = "confirmed"
)

func ParsePlanningState(raw string) PlanningState {
	switch PlanningState(raw) {
	case StateIdle, StateClarifying, StateReview, StateConfirmed:
		return PlanningState(raw)
	default:
		return StateIdle
	}
}

func (s PlanningState) IsFinal() bool {
	return s == StateConfirmed
}

const (
	IntentClarifyGoal     = "clarify_goal"
	IntentConfirmPlan     = "confirm_plan"
	IntentFallbackUnknown = "fallback_unknown"
)

type IntentResult struct {
	Intent     string
	Confidence float64
}

const (
	ConversationRoleUser      = "user"
	ConversationRoleAssistant = "assistant"
	ConversationRoleSystem    = "system"
)

const (
	SlotMainGoal        = "main_goal"
	SlotSuccessCriteria = "success_criteria"
	SlotCurrentLevel    = "current_level"
	SlotTimeBudget      = "time_budget"
	SlotConstraints     = "constraints"
	SlotRiskFlags       = "risk_flags"
)

var requiredSlotOrder = []string{
	SlotMainGoal,
	SlotSuccessCriteria,
	SlotCurrentLevel,
	SlotTimeBudget,
	SlotConstraints,
	SlotRiskFlags,
}

func DefaultSlotCompletion() map[string]bool {
	result := make(map[string]bool, len(requiredSlotOrder))
	for _, key := range requiredSlotOrder {
		result[key] = false
	}
	return result
}

func NormalizeSlotCompletion(input map[string]bool) map[string]bool {
	result := DefaultSlotCompletion()
	for key, value := range input {
		if _, ok := result[key]; ok {
			result[key] = value
		}
	}
	return result
}

func MissingRequiredSlots(slotCompletion map[string]bool) []string {
	normalized := NormalizeSlotCompletion(slotCompletion)
	missing := make([]string, 0, len(requiredSlotOrder))
	for _, key := range requiredSlotOrder {
		if !normalized[key] {
			missing = append(missing, key)
		}
	}
	return missing
}

func IsRequiredSlotsComplete(slotCompletion map[string]bool) bool {
	return len(MissingRequiredSlots(slotCompletion)) == 0
}

func SlotLabel(slotKey string) string {
	switch slotKey {
	case SlotMainGoal:
		return "主目标"
	case SlotSuccessCriteria:
		return "成功标准"
	case SlotCurrentLevel:
		return "当前水平"
	case SlotTimeBudget:
		return "时间预算"
	case SlotConstraints:
		return "约束条件"
	case SlotRiskFlags:
		return "风险项"
	default:
		return slotKey
	}
}
