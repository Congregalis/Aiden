package telegram

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestIntentRouterRoutesExpectedIntents(t *testing.T) {
	router := NewIntentRouter()

	if got := router.Route("/goal", StateIdle); got.Intent != IntentClarifyGoal {
		t.Fatalf("intent for /goal=%q, want %q", got.Intent, IntentClarifyGoal)
	}

	if got := router.Route("确认", StateReview); got.Intent != IntentConfirmPlan {
		t.Fatalf("intent for confirm=%q, want %q", got.Intent, IntentConfirmPlan)
	}

	if got := router.Route("嗯", StateClarifying); got.Intent != IntentFallbackUnknown {
		t.Fatalf("intent for fallback=%q, want %q", got.Intent, IntentFallbackUnknown)
	}
}

func TestWorkerGoalMovesSessionToClarifyingAndSavesTurns(t *testing.T) {
	store := newMemoryStore()
	client := &scriptedClient{
		updates: [][]Update{{
			{UpdateID: 1, Message: &Message{MessageID: 21, Chat: Chat{ID: 20001}, Text: "/goal"}},
		}},
	}

	if err := runWorkerUntilSendCount(t, client, store, 1); err != nil {
		t.Fatalf("worker run failed: %v", err)
	}

	user, ok := store.UserByChatID(20001)
	if !ok {
		t.Fatalf("expected user to exist")
	}
	goal, found, err := store.GetActiveGoalByUserID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("get active goal failed: %v", err)
	}
	if !found {
		t.Fatalf("expected active goal")
	}

	session, ok := store.SessionByGoalID(goal.ID)
	if !ok {
		t.Fatalf("expected planning session")
	}
	if session.State != StateClarifying {
		t.Fatalf("session state=%q, want %q", session.State, StateClarifying)
	}
	if session.TurnCount != 1 {
		t.Fatalf("turn_count=%d, want 1", session.TurnCount)
	}

	turns := store.ConversationTurnsBySessionID(session.ID)
	if len(turns) != 2 {
		t.Fatalf("conversation turns=%d, want 2", len(turns))
	}
	if turns[0].Role != ConversationRoleUser || turns[1].Role != ConversationRoleAssistant {
		t.Fatalf("unexpected turn roles: %q, %q", turns[0].Role, turns[1].Role)
	}
}

func TestWorkerMovesToReviewWhenRequiredSlotsComplete(t *testing.T) {
	store := newMemoryStore()
	client := &scriptedClient{
		updates: [][]Update{{
			{
				UpdateID: 1,
				Message: &Message{
					MessageID: 22,
					Chat:      Chat{ID: 20002},
					Text:      "我想在3个月内通过Go面试，成功标准是1.完成3个项目 2.刷100题 3.通过面试，我是零基础，每周10小时，工作日晚上学习，限制是经常加班，风险是容易拖延。",
				},
			},
		}},
	}

	if err := runWorkerUntilSendCount(t, client, store, 1); err != nil {
		t.Fatalf("worker run failed: %v", err)
	}

	user, _ := store.UserByChatID(20002)
	goal, _, _ := store.GetActiveGoalByUserID(context.Background(), user.ID)
	session, _ := store.SessionByGoalID(goal.ID)
	if session.State != StateReview {
		t.Fatalf("session state=%q, want %q", session.State, StateReview)
	}

	sent := client.SentMessages()
	if len(sent) != 1 || !strings.Contains(sent[0].Text, ReplyReviewReady) {
		t.Fatalf("reply=%q, want review ready", sent[0].Text)
	}
}

func TestWorkerReviewConfirmationMovesToConfirmed(t *testing.T) {
	store := newMemoryStore()
	client := &scriptedClient{
		updates: [][]Update{{
			{
				UpdateID: 1,
				Message: &Message{
					MessageID: 23,
					Chat:      Chat{ID: 20003},
					Text:      "我想在3个月内通过Go面试，成功标准是1.完成3个项目 2.刷100题 3.通过面试，我是零基础，每周10小时，工作日晚上学习，限制是经常加班，风险是容易拖延。",
				},
			},
			{UpdateID: 2, Message: &Message{MessageID: 24, Chat: Chat{ID: 20003}, Text: "确认"}},
		}},
	}

	if err := runWorkerUntilSendCount(t, client, store, 2); err != nil {
		t.Fatalf("worker run failed: %v", err)
	}

	user, _ := store.UserByChatID(20003)
	goal, _, _ := store.GetActiveGoalByUserID(context.Background(), user.ID)
	session, _ := store.SessionByGoalID(goal.ID)
	if session.State != StateConfirmed {
		t.Fatalf("session state=%q, want %q", session.State, StateConfirmed)
	}

	sent := client.SentMessages()
	if got := sent[len(sent)-1].Text; got != ReplyPlanConfirmed {
		t.Fatalf("confirm reply=%q, want %q", got, ReplyPlanConfirmed)
	}
}

func TestWorkerReviewModificationReturnsToClarifying(t *testing.T) {
	store := newMemoryStore()
	client := &scriptedClient{
		updates: [][]Update{{
			{
				UpdateID: 1,
				Message: &Message{
					MessageID: 25,
					Chat:      Chat{ID: 20004},
					Text:      "我想在3个月内通过Go面试，成功标准是1.完成3个项目 2.刷100题 3.通过面试，我是零基础，每周10小时，工作日晚上学习，限制是经常加班，风险是容易拖延。",
				},
			},
			{UpdateID: 2, Message: &Message{MessageID: 26, Chat: Chat{ID: 20004}, Text: "我想修改成每周6小时，继续优化"}},
		}},
	}

	if err := runWorkerUntilSendCount(t, client, store, 2); err != nil {
		t.Fatalf("worker run failed: %v", err)
	}

	user, _ := store.UserByChatID(20004)
	goal, _, _ := store.GetActiveGoalByUserID(context.Background(), user.ID)
	session, _ := store.SessionByGoalID(goal.ID)
	if session.State != StateClarifying {
		t.Fatalf("session state=%q, want %q", session.State, StateClarifying)
	}

	sent := client.SentMessages()
	if got := sent[len(sent)-1].Text; !strings.Contains(got, "已收到修改意见") {
		t.Fatalf("reply=%q, want modification acknowledgement", got)
	}
}

func TestWorkerLimitsFollowUpQuestionsToTwo(t *testing.T) {
	store := newMemoryStore()
	client := &scriptedClient{
		updates: [][]Update{{
			{UpdateID: 1, Message: &Message{MessageID: 27, Chat: Chat{ID: 20005}, Text: "我想学Go"}},
		}},
	}

	if err := runWorkerUntilSendCount(t, client, store, 1); err != nil {
		t.Fatalf("worker run failed: %v", err)
	}

	reply := client.SentMessages()[0].Text
	re := regexp.MustCompile(`\n[1-9]\)`)
	if count := len(re.FindAllString(reply, -1)); count > 2 {
		t.Fatalf("follow-up question count=%d, want <=2, reply=%q", count, reply)
	}
}

func TestWorkerAddsSummaryEveryThreeTurns(t *testing.T) {
	store := newMemoryStore()
	client := &scriptedClient{
		updates: [][]Update{{
			{UpdateID: 1, Message: &Message{MessageID: 28, Chat: Chat{ID: 20006}, Text: "/goal"}},
			{UpdateID: 2, Message: &Message{MessageID: 29, Chat: Chat{ID: 20006}, Text: "我是零基础"}},
			{UpdateID: 3, Message: &Message{MessageID: 30, Chat: Chat{ID: 20006}, Text: "每周5小时，工作日晚上学习"}},
		}},
	}

	if err := runWorkerUntilSendCount(t, client, store, 3); err != nil {
		t.Fatalf("worker run failed: %v", err)
	}

	sent := client.SentMessages()
	third := sent[2].Text
	if !strings.Contains(third, "【当前摘要】") {
		t.Fatalf("third reply=%q, want progress summary", third)
	}
	if !strings.Contains(third, "继续优化") {
		t.Fatalf("third reply=%q, want continue optimize prompt", third)
	}
}

func TestWorkerTimeoutResetsSessionWithReminder(t *testing.T) {
	store := newMemoryStore()
	user, _, err := store.FindOrCreateUserByChatID(context.Background(), 20007)
	if err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	goal, err := store.CreateGoalDraft(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("seed goal failed: %v", err)
	}
	session, _, err := store.GetOrCreatePlanningSession(context.Background(), goal.ID)
	if err != nil {
		t.Fatalf("seed session failed: %v", err)
	}
	session.State = StateReview
	session.UpdatedAt = time.Now().Add(-25 * time.Hour)
	if err := store.UpdatePlanningSession(context.Background(), session); err != nil {
		t.Fatalf("seed update session failed: %v", err)
	}
	store.mu.Lock()
	seeded := store.sessionsByGoalID[goal.ID]
	seeded.UpdatedAt = time.Now().Add(-25 * time.Hour)
	store.sessionsByGoalID[goal.ID] = seeded
	store.mu.Unlock()

	client := &scriptedClient{
		updates: [][]Update{{
			{UpdateID: 1, Message: &Message{MessageID: 31, Chat: Chat{ID: 20007}, Text: "继续"}},
		}},
	}

	if err := runWorkerUntilSendCount(t, client, store, 1); err != nil {
		t.Fatalf("worker run failed: %v", err)
	}

	reply := client.SentMessages()[0].Text
	if !strings.Contains(reply, ReplySessionTimeout) {
		t.Fatalf("reply=%q, want timeout reminder", reply)
	}

	updated, _ := store.SessionByGoalID(goal.ID)
	if updated.State != StateClarifying {
		t.Fatalf("session state=%q, want %q", updated.State, StateClarifying)
	}
}

func TestFallbackIntentKeepsContext(t *testing.T) {
	store := newMemoryStore()
	client := &scriptedClient{
		updates: [][]Update{{
			{UpdateID: 1, Message: &Message{MessageID: 32, Chat: Chat{ID: 20008}, Text: "我想3个月学会Go并完成项目"}},
			{UpdateID: 2, Message: &Message{MessageID: 33, Chat: Chat{ID: 20008}, Text: "嗯"}},
		}},
	}

	if err := runWorkerUntilSendCount(t, client, store, 2); err != nil {
		t.Fatalf("worker run failed: %v", err)
	}

	user, _ := store.UserByChatID(20008)
	goal, _, _ := store.GetActiveGoalByUserID(context.Background(), user.ID)
	session, _ := store.SessionByGoalID(goal.ID)
	if session.State != StateClarifying {
		t.Fatalf("session state=%q, want %q", session.State, StateClarifying)
	}
	if !session.SlotCompletion[SlotMainGoal] {
		t.Fatalf("expected main_goal to remain completed")
	}

	secondReply := client.SentMessages()[1].Text
	if secondReply != ReplyFallbackGuidance {
		t.Fatalf("fallback reply=%q, want %q", secondReply, ReplyFallbackGuidance)
	}
}
