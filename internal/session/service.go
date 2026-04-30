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

type OnSessionEndFunc func(ctx context.Context, sessionID string)

type Service struct {
	store        *store.SessionStore
	memory       *memory.Manager
	llmClient    llm.Client
	logger       *slog.Logger
	OnSessionEnd OnSessionEndFunc
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

type EndSessionResult struct {
	SessionID  string
	Status     string
	FinalRound int
}

type SessionDetail struct {
	SessionID       string
	Status          string
	RoleDescription string
	RoundLimit      int
	CurrentRound    int
	MessageCount    int
	CreatedAt       time.Time
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
	rc, goals, dims, err := s.parseSessionConfig(sess)
	if err != nil {
		return nil, fmt.Errorf("parse session config: %w", err)
	}

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
		s.notifySessionEnd(sessionID)
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

func (s *Service) EndSession(ctx context.Context, sessionID string) (*EndSessionResult, error) {
	sess, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if sess == nil {
		return nil, ErrSessionNotFound
	}
	if sess.Status != "active" {
		if sess.Status == "completed" {
			msgCount, _ := s.store.CountMessages(ctx, sessionID)
			return &EndSessionResult{
				SessionID:  sessionID,
				Status:     "completed",
				FinalRound: msgCount / 2,
			}, nil
		}
		return nil, ErrSessionCompleted
	}

	if err := s.store.UpdateSessionStatus(ctx, sessionID, "completed"); err != nil {
		return nil, fmt.Errorf("end session: %w", err)
	}

	msgCount, err := s.store.CountMessages(ctx, sessionID)
	if err != nil {
		s.logger.Warn("count messages after end session failed", "session_id", sessionID, "error", err)
		msgCount = 0
	}

	s.logger.Info("session ended manually", "session_id", sessionID, "final_round", msgCount/2, "trigger", "manual")
	s.notifySessionEnd(sessionID)
	return &EndSessionResult{
		SessionID:  sessionID,
		Status:     "completed",
		FinalRound: msgCount / 2,
	}, nil
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

func (s *Service) GetSessionDetail(ctx context.Context, sessionID string) (*SessionDetail, error) {
	sess, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if sess == nil {
		return nil, ErrSessionNotFound
	}

	rc, _, _, err := s.parseSessionConfig(sess)
	if err != nil {
		return nil, fmt.Errorf("parse session config: %w", err)
	}

	msgCount, err := s.store.CountMessages(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("count messages: %w", err)
	}

	return &SessionDetail{
		SessionID:       sess.ID,
		Status:          sess.Status,
		RoleDescription: rc.Description,
		RoundLimit:      sess.RoundLimit,
		CurrentRound:    msgCount / 2,
		MessageCount:    msgCount,
		CreatedAt:       sess.CreatedAt,
	}, nil
}

func (s *Service) notifySessionEnd(sessionID string) {
	if s.OnSessionEnd != nil {
		ctx := context.Background()
		s.OnSessionEnd(ctx, sessionID)
	}
}

func (s *Service) parseSessionConfig(sess *store.Session) (roleConfigJSON, []role.TrainingGoal, []role.Dimension, error) {
	var rc roleConfigJSON
	if err := json.Unmarshal([]byte(sess.RoleConfig), &rc); err != nil {
		return rc, nil, nil, fmt.Errorf("unmarshal role config: %w", err)
	}
	var goals []role.TrainingGoal
	if err := json.Unmarshal([]byte(sess.Goals), &goals); err != nil {
		return rc, nil, nil, fmt.Errorf("unmarshal goals: %w", err)
	}
	var dims []role.Dimension
	if err := json.Unmarshal([]byte(sess.Dimensions), &dims); err != nil {
		return rc, nil, nil, fmt.Errorf("unmarshal dimensions: %w", err)
	}
	return rc, goals, dims, nil
}

type roleConfigJSON struct {
	Description string        `json:"description"`
	Scenario    string        `json:"scenario"`
	Type        role.RoleType `json:"type"`
}

func BuildSystemPrompt(desc, scenario string, goals []role.TrainingGoal, dims []role.Dimension) string {
	var b strings.Builder
	fmt.Fprintf(&b, "你将扮演以下角色与用户进行对话：\n\n你的角色：%s\n场景背景：%s\n\n请完全沉浸在这个角色中，用该角色的身份、语气和知识背景与用户对话。用户的目标是在这个场景中练习自己的应对和表达能力。", desc, scenario)

	if len(goals) > 0 {
		b.WriteString("\n\n训练目标（在这些方面帮助用户提升）：")
		for _, g := range goals {
			fmt.Fprintf(&b, "\n- %s：%s", g.Name, g.Description)
		}
	}

	if len(dims) > 0 {
		b.WriteString("\n\n对话结束后，将从以下维度分析用户的表现：")
		for _, d := range dims {
			fmt.Fprintf(&b, "\n- %s：%s", d.Name, d.Description)
		}
	}

	b.WriteString("\n\n请自然地进行对话，始终保持在角色中，引导对话朝训练目标发展。")
	return b.String()
}
