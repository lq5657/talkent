package session

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lq5657/talkent/internal/llm"
	"github.com/lq5657/talkent/internal/memory"
	"github.com/lq5657/talkent/internal/role"
	"github.com/lq5657/talkent/internal/store"
)

type Service struct {
	store     *store.SessionStore
	memory    *memory.Manager
	llmClient llm.Client
	logger    *slog.Logger
}

func NewService(s *store.SessionStore, m *memory.Manager, llmClient llm.Client, logger *slog.Logger) *Service {
	return &Service{
		store:     s,
		memory:    m,
		llmClient: llmClient,
		logger:    logger,
	}
}

type CreateSessionRequest struct {
	RoleDescription string             `json:"role_description"`
	Scenario        string             `json:"scenario"`
	RoleType        role.RoleType      `json:"role_type"`
	Goals           []role.TrainingGoal `json:"goals"`
	Dimensions      []role.Dimension   `json:"dimensions"`
	RoundLimit      int                `json:"round_limit"`
}

type ChatResult struct {
	Reply        string `json:"reply"`
	CurrentRound int    `json:"current_round"`
	RoundLimit   int    `json:"limit"`
	IsLast       bool   `json:"is_last"`
	MemorySource string `json:"memory_source"`
}

func (s *Service) CreateSession(ctx context.Context, req CreateSessionRequest) (*store.Session, error) {
	now := time.Now()
	roleConfig := roleConfigJSON{Description: req.RoleDescription, Scenario: req.Scenario, Type: req.RoleType}
	roleConfigBytes, err := json.Marshal(roleConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal role config: %w", err)
	}

	goalsBytes, err := json.Marshal(req.Goals)
	if err != nil {
		return nil, fmt.Errorf("marshal goals: %w", err)
	}

	dimsBytes, err := json.Marshal(req.Dimensions)
	if err != nil {
		return nil, fmt.Errorf("marshal dimensions: %w", err)
	}

	session := &store.Session{
		ID:         uuid.New().String(),
		RoleConfig: string(roleConfigBytes),
		Goals:      string(goalsBytes),
		Dimensions: string(dimsBytes),
		Status:     "active",
		RoundLimit: req.RoundLimit,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.store.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	s.logger.Info("session created", "session_id", session.ID, "role_type", req.RoleType)
	return session, nil
}

func (s *Service) Chat(ctx context.Context, sessionID string, userContent string) (*ChatResult, error) {
	sess, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if sess == nil {
		return nil, ErrSessionNotFound
	}
	if sess.Status != "active" {
		return nil, ErrSessionCompleted
	}

	msgCount, err := s.store.CountMessages(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("count messages: %w", err)
	}
	currentRound := msgCount / 2

	// Persist user message
	userMsg := &store.Message{
		SessionID:   sessionID,
		Role:        "user",
		Content:     userContent,
		SequenceNum: msgCount + 1,
		CreatedAt:   time.Now(),
	}
	if err := s.store.CreateMessage(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("create user message: %w", err)
	}

	// Build context with memory
	var rc roleConfigJSON
	json.Unmarshal([]byte(sess.RoleConfig), &rc)
	var goals []role.TrainingGoal
	json.Unmarshal([]byte(sess.Goals), &goals)
	var dims []role.Dimension
	json.Unmarshal([]byte(sess.Dimensions), &dims)

	systemPrompt := BuildSystemPrompt(rc.Description, rc.Scenario, goals, dims)

	history, err := s.store.ListMessages(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}

	ctxResult, err := s.memory.BuildContext(ctx, systemPrompt, history)
	if err != nil {
		return nil, fmt.Errorf("build context: %w", err)
	}

	// Call LLM
	resp, err := s.llmClient.Chat(ctx, ctxResult.Messages, &llm.ChatOptions{Temperature: 0.7})
	if err != nil {
		return nil, fmt.Errorf("llm chat: %w", err)
	}

	// Persist assistant message
	assistantMsg := &store.Message{
		SessionID:   sessionID,
		Role:        "assistant",
		Content:     resp.Content,
		SequenceNum: msgCount + 2,
		CreatedAt:   time.Now(),
	}
	if err := s.store.CreateMessage(ctx, assistantMsg); err != nil {
		return nil, fmt.Errorf("create assistant message: %w", err)
	}

	currentRound++
	isLast := sess.RoundLimit > 0 && currentRound >= sess.RoundLimit

	if isLast {
		if err := s.store.UpdateSessionStatus(ctx, sessionID, "completed"); err != nil {
			return nil, fmt.Errorf("auto end session: %w", err)
		}
		s.logger.Info("session auto-ended", "session_id", sessionID, "final_round", currentRound, "trigger", "auto")
	}

	s.logger.Info("chat round completed", "session_id", sessionID, "round", currentRound, "memory_source", ctxResult.MemorySource)

	return &ChatResult{
		Reply:        resp.Content,
		CurrentRound: currentRound,
		RoundLimit:   sess.RoundLimit,
		IsLast:       isLast,
		MemorySource: ctxResult.MemorySource,
	}, nil
}

func (s *Service) EndSession(ctx context.Context, sessionID string) (*store.Session, error) {
	sess, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if sess == nil {
		return nil, ErrSessionNotFound
	}
	if sess.Status != "active" {
		return nil, ErrSessionCompleted
	}

	if err := s.store.UpdateSessionStatus(ctx, sessionID, "completed"); err != nil {
		return nil, fmt.Errorf("end session: %w", err)
	}

	sess.Status = "completed"
	msgCount, _ := s.store.CountMessages(ctx, sessionID)
	s.logger.Info("session ended manually", "session_id", sessionID, "final_round", msgCount/2, "trigger", "manual")
	return sess, nil
}

func (s *Service) GetSession(ctx context.Context, sessionID string) (*store.Session, error) {
	sess, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if sess == nil {
		return nil, ErrSessionNotFound
	}
	return sess, nil
}

type roleConfigJSON struct {
	Description string        `json:"description"`
	Scenario    string        `json:"scenario"`
	Type        role.RoleType `json:"type"`
}

func BuildSystemPrompt(desc, scenario string, goals []role.TrainingGoal, dims []role.Dimension) string {
	var b strings.Builder
	fmt.Fprintf(&b, "你是一位对话训练伙伴。请按照以下设定进行对话：\n\n角色设定：%s\n场景：%s", desc, scenario)

	if len(goals) > 0 {
		b.WriteString("\n训练目标：")
		for _, g := range goals {
			fmt.Fprintf(&b, "\n- %s：%s", g.Name, g.Description)
		}
	}

	if len(dims) > 0 {
		b.WriteString("\n期望分析维度：")
		for _, d := range dims {
			fmt.Fprintf(&b, "\n- %s：%s", d.Name, d.Description)
		}
	}

	b.WriteString("\n\n请自然地进行对话，保持角色设定，引导对话朝训练目标发展。")
	return b.String()
}
