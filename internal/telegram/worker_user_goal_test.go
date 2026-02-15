package telegram

import (
	"context"
	"testing"
)

func TestWorkerStartInitializesNewUser(t *testing.T) {
	store := newMemoryStore()
	client := &scriptedClient{
		updates: [][]Update{{
			{UpdateID: 1, Message: &Message{MessageID: 11, Chat: Chat{ID: 10001}, Text: "/start"}},
		}},
	}

	if err := runWorkerUntilSendCount(t, client, store, 1); err != nil {
		t.Fatalf("worker run failed: %v", err)
	}

	createdUser, ok := store.UserByChatID(10001)
	if !ok {
		t.Fatalf("expected user to be created")
	}
	if createdUser.Language != "zh-CN" {
		t.Fatalf("language=%q, want zh-CN", createdUser.Language)
	}
	if createdUser.Timezone != "Asia/Shanghai" {
		t.Fatalf("timezone=%q, want Asia/Shanghai", createdUser.Timezone)
	}

	sent := client.SentMessages()
	if len(sent) != 1 {
		t.Fatalf("sent messages=%d, want 1", len(sent))
	}
	if sent[0].Text != ReplyStart {
		t.Fatalf("reply=%q, want %q", sent[0].Text, ReplyStart)
	}
}

func TestWorkerStartWelcomesBackExistingUser(t *testing.T) {
	store := newMemoryStore()
	_, _, err := store.FindOrCreateUserByChatID(context.Background(), 10002)
	if err != nil {
		t.Fatalf("seed user failed: %v", err)
	}

	client := &scriptedClient{
		updates: [][]Update{{
			{UpdateID: 1, Message: &Message{MessageID: 12, Chat: Chat{ID: 10002}, Text: "/start"}},
		}},
	}

	if err := runWorkerUntilSendCount(t, client, store, 1); err != nil {
		t.Fatalf("worker run failed: %v", err)
	}

	sent := client.SentMessages()
	if len(sent) != 1 {
		t.Fatalf("sent messages=%d, want 1", len(sent))
	}
	if sent[0].Text != ReplyStartBack {
		t.Fatalf("reply=%q, want %q", sent[0].Text, ReplyStartBack)
	}
}

func TestWorkerGoalCreatesDraftWhenNoActiveGoal(t *testing.T) {
	store := newMemoryStore()
	client := &scriptedClient{
		updates: [][]Update{{
			{UpdateID: 1, Message: &Message{MessageID: 13, Chat: Chat{ID: 10003}, Text: "/goal"}},
		}},
	}

	if err := runWorkerUntilSendCount(t, client, store, 1); err != nil {
		t.Fatalf("worker run failed: %v", err)
	}

	if store.GoalCreateCount() != 1 {
		t.Fatalf("goal create count=%d, want 1", store.GoalCreateCount())
	}

	createdUser, ok := store.UserByChatID(10003)
	if !ok {
		t.Fatalf("expected user to exist")
	}

	goal, found, err := store.GetActiveGoalByUserID(context.Background(), createdUser.ID)
	if err != nil {
		t.Fatalf("get active goal failed: %v", err)
	}
	if !found {
		t.Fatalf("expected active goal to exist")
	}
	if goal.Status != "draft" {
		t.Fatalf("goal status=%q, want draft", goal.Status)
	}

	sent := client.SentMessages()
	if len(sent) != 1 {
		t.Fatalf("sent messages=%d, want 1", len(sent))
	}
	if sent[0].Text != ReplyGoal {
		t.Fatalf("reply=%q, want %q", sent[0].Text, ReplyGoal)
	}
}

func TestWorkerGoalReusesExistingActiveGoal(t *testing.T) {
	store := newMemoryStore()
	user, _, err := store.FindOrCreateUserByChatID(context.Background(), 10004)
	if err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	if _, err := store.CreateGoalDraft(context.Background(), user.ID); err != nil {
		t.Fatalf("seed goal failed: %v", err)
	}

	client := &scriptedClient{
		updates: [][]Update{{
			{UpdateID: 1, Message: &Message{MessageID: 14, Chat: Chat{ID: 10004}, Text: "/goal"}},
		}},
	}

	if err := runWorkerUntilSendCount(t, client, store, 1); err != nil {
		t.Fatalf("worker run failed: %v", err)
	}

	if store.GoalCreateCount() != 1 {
		t.Fatalf("goal create count=%d, want 1", store.GoalCreateCount())
	}
}

func TestWorkerNaturalMessageCreatesDraftWhenNoActiveGoal(t *testing.T) {
	store := newMemoryStore()
	client := &scriptedClient{
		updates: [][]Update{{
			{UpdateID: 1, Message: &Message{MessageID: 15, Chat: Chat{ID: 10005}, Text: "我想在三个月内学完 Go"}},
		}},
	}

	if err := runWorkerUntilSendCount(t, client, store, 1); err != nil {
		t.Fatalf("worker run failed: %v", err)
	}

	if store.GoalCreateCount() != 1 {
		t.Fatalf("goal create count=%d, want 1", store.GoalCreateCount())
	}

	sent := client.SentMessages()
	if len(sent) != 1 {
		t.Fatalf("sent messages=%d, want 1", len(sent))
	}
	if sent[0].Text != ReplyNaturalMessage {
		t.Fatalf("reply=%q, want %q", sent[0].Text, ReplyNaturalMessage)
	}
}
