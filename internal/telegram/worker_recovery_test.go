package telegram

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"
)

func TestWorkerRestartRecoverySkipsDuplicateMessages(t *testing.T) {
	store := newMemoryStore()

	firstRunClient := &scriptedClient{
		updates: [][]Update{{
			{UpdateID: 100, Message: &Message{MessageID: 1, Chat: Chat{ID: 12345678}, Text: "/help"}},
		}},
	}

	if err := runWorkerUntilSendCount(t, firstRunClient, store, 1); err != nil {
		t.Fatalf("first run failed: %v", err)
	}

	if store.LastUpdateID() != 100 {
		t.Fatalf("first run last update id=%d, want 100", store.LastUpdateID())
	}

	secondRunClient := &scriptedClient{
		updates: [][]Update{{
			{UpdateID: 100, Message: &Message{MessageID: 2, Chat: Chat{ID: 12345678}, Text: "/help"}},
			{UpdateID: 101, Message: &Message{MessageID: 3, Chat: Chat{ID: 12345678}, Text: "/goal"}},
		}},
	}

	if err := runWorkerUntilSendCount(t, secondRunClient, store, 1); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	offsets := secondRunClient.GetOffsets()
	if len(offsets) == 0 {
		t.Fatalf("expected getUpdates to be called")
	}
	if offsets[0] != 101 {
		t.Fatalf("first offset on restart=%d, want 101", offsets[0])
	}

	if secondRunClient.SendCount() != 1 {
		t.Fatalf("second run send count=%d, want 1", secondRunClient.SendCount())
	}

	if store.LastUpdateID() != 101 {
		t.Fatalf("second run last update id=%d, want 101", store.LastUpdateID())
	}
}

func runWorkerUntilSendCount(t *testing.T, client *scriptedClient, store *memoryStore, sendCount int) error {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	worker := NewWorker(WorkerConfig{
		PollTimeoutSec: 1,
		PollInterval:   5 * time.Millisecond,
		AllowedUpdates: []string{"message"},
	}, client, store, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- worker.Run(ctx)
	}()

	deadline := time.After(2 * time.Second)
	for client.SendCount() < sendCount {
		select {
		case <-deadline:
			cancel()
			<-errCh
			return context.DeadlineExceeded
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}

	cancel()
	return <-errCh
}

type memoryStore struct {
	mu               sync.Mutex
	lastUpdateID     int64
	dedup            map[int64]struct{}
	usersByChatID    map[int64]User
	activeGoalByUID  map[string]Goal
	sessionsByGoalID map[string]PlanningSession
	turns            []ConversationTurn
	nextUserID       int
	nextGoalID       int
	nextSessionID    int
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		dedup:            make(map[int64]struct{}),
		usersByChatID:    make(map[int64]User),
		activeGoalByUID:  make(map[string]Goal),
		sessionsByGoalID: make(map[string]PlanningSession),
		turns:            make([]ConversationTurn, 0),
	}
}

func (s *memoryStore) LoadLastUpdateID(context.Context) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastUpdateID, nil
}

func (s *memoryStore) SaveLastUpdateID(_ context.Context, lastUpdateID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastUpdateID = lastUpdateID
	return nil
}

func (s *memoryStore) MarkMessageDedup(_ context.Context, updateID, _ int64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.dedup[updateID]; exists {
		return false, nil
	}
	s.dedup[updateID] = struct{}{}
	return true, nil
}

func (s *memoryStore) FindOrCreateUserByChatID(_ context.Context, chatID int64) (User, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.usersByChatID[chatID]; ok {
		return existing, false, nil
	}

	s.nextUserID++
	created := User{
		ID:             fmt.Sprintf("user-%d", s.nextUserID),
		TelegramChatID: chatID,
		Language:       "zh-CN",
		Timezone:       "Asia/Shanghai",
	}
	s.usersByChatID[chatID] = created
	return created, true, nil
}

func (s *memoryStore) GetActiveGoalByUserID(_ context.Context, userID string) (Goal, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	goal, ok := s.activeGoalByUID[userID]
	if !ok {
		return Goal{}, false, nil
	}

	return goal, true, nil
}

func (s *memoryStore) CreateGoalDraft(_ context.Context, userID string) (Goal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextGoalID++
	goal := Goal{
		ID:     fmt.Sprintf("goal-%d", s.nextGoalID),
		UserID: userID,
		Title:  "",
		Status: "draft",
	}
	s.activeGoalByUID[userID] = goal
	return goal, nil
}

func (s *memoryStore) GetOrCreatePlanningSession(_ context.Context, goalID string) (PlanningSession, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.sessionsByGoalID[goalID]; ok {
		return existing, false, nil
	}

	s.nextSessionID++
	created := PlanningSession{
		ID:             fmt.Sprintf("session-%d", s.nextSessionID),
		GoalID:         goalID,
		State:          StateIdle,
		SlotCompletion: DefaultSlotCompletion(),
		TurnCount:      0,
		LastIntent:     "",
		UpdatedAt:      time.Now(),
	}
	s.sessionsByGoalID[goalID] = created
	return created, true, nil
}

func (s *memoryStore) IncrementPlanningSessionTurn(_ context.Context, sessionID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, key, ok := s.findSessionByID(sessionID)
	if !ok {
		return 0, fmt.Errorf("planning session %s not found", sessionID)
	}

	session.TurnCount++
	session.UpdatedAt = time.Now()
	s.sessionsByGoalID[key] = session
	return session.TurnCount, nil
}

func (s *memoryStore) UpdatePlanningSession(_ context.Context, updated PlanningSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessionsByGoalID[updated.GoalID]; !ok {
		return fmt.Errorf("planning session %s not found", updated.ID)
	}

	updated.SlotCompletion = NormalizeSlotCompletion(updated.SlotCompletion)
	updated.UpdatedAt = time.Now()
	s.sessionsByGoalID[updated.GoalID] = updated
	return nil
}

func (s *memoryStore) SaveConversationTurn(_ context.Context, turn ConversationTurn) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.turns = append(s.turns, turn)
	return nil
}

func (s *memoryStore) LastUpdateID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastUpdateID
}

func (s *memoryStore) GoalCreateCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.nextGoalID
}

func (s *memoryStore) UserByChatID(chatID int64) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.usersByChatID[chatID]
	return user, ok
}

func (s *memoryStore) SessionByGoalID(goalID string) (PlanningSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessionsByGoalID[goalID]
	return session, ok
}

func (s *memoryStore) ConversationTurnsBySessionID(sessionID string) []ConversationTurn {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ConversationTurn, 0, len(s.turns))
	for _, turn := range s.turns {
		if turn.SessionID == sessionID {
			out = append(out, turn)
		}
	}
	return out
}

func (s *memoryStore) findSessionByID(sessionID string) (PlanningSession, string, bool) {
	for goalID, session := range s.sessionsByGoalID {
		if session.ID == sessionID {
			return session, goalID, true
		}
	}
	return PlanningSession{}, "", false
}

type scriptedClient struct {
	mu      sync.Mutex
	updates [][]Update
	offsets []int64
	sent    []OutgoingMessage
}

func (c *scriptedClient) GetMe(context.Context) (BotUser, error) {
	return BotUser{ID: 1, Username: "aiden_test_bot"}, nil
}

func (c *scriptedClient) GetUpdates(ctx context.Context, params GetUpdatesParams) ([]Update, error) {
	c.mu.Lock()
	c.offsets = append(c.offsets, params.Offset)
	if len(c.updates) > 0 {
		batch := c.updates[0]
		c.updates = c.updates[1:]
		c.mu.Unlock()
		return batch, nil
	}
	c.mu.Unlock()

	<-ctx.Done()
	return nil, nil
}

func (c *scriptedClient) SendMessage(_ context.Context, message OutgoingMessage) (Message, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sent = append(c.sent, message)
	return Message{}, nil
}

func (c *scriptedClient) SendCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.sent)
}

func (c *scriptedClient) GetOffsets() []int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]int64, len(c.offsets))
	copy(out, c.offsets)
	return out
}

func (c *scriptedClient) SentMessages() []OutgoingMessage {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]OutgoingMessage, len(c.sent))
	copy(out, c.sent)
	return out
}
