package memory

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lq5657/talkent/internal/llm"
	"github.com/lq5657/talkent/internal/store"
)

const summarizeSystemPrompt = `你是一位对话摘要专家。请对以下对话历史生成简洁的结构化摘要，保留关键信息、立场和上下文。
你必须以纯文本形式输出摘要，不要输出 JSON 或其他格式。`

type Manager struct {
	windowSize int
	llmClient  llm.Client
	logger     *slog.Logger
}

func NewManager(windowSize int, llmClient llm.Client, logger *slog.Logger) *Manager {
	return &Manager{
		windowSize: windowSize,
		llmClient:  llmClient,
		logger:     logger,
	}
}

type BuildContextResult struct {
	Messages     []llm.ChatMessage
	MemorySource string // "window" or "summary+window"
}

func (m *Manager) BuildContext(ctx context.Context, systemPrompt string, messages []store.Message) (*BuildContextResult, error) {
	total := len(messages)
	if total <= m.windowSize {
		chatMsgs := m.buildWindowMessages(systemPrompt, messages)
		return &BuildContextResult{Messages: chatMsgs, MemorySource: "window"}, nil
	}

	overflow := messages[:total-m.windowSize]
	window := messages[total-m.windowSize:]

	summary, err := m.Summarize(ctx, overflow)
	if err != nil {
		m.logger.Warn("summary failed, degrading to window-only", "error", err, "overflow_count", len(overflow))
		chatMsgs := m.buildWindowMessages(systemPrompt, window)
		return &BuildContextResult{Messages: chatMsgs, MemorySource: "window"}, nil
	}

	enrichedPrompt := systemPrompt + "\n\n## 对话历史摘要\n" + summary
	chatMsgs := m.buildWindowMessages(enrichedPrompt, window)
	m.logger.Info("summary generated for overflow", "overflow_count", len(overflow))
	return &BuildContextResult{Messages: chatMsgs, MemorySource: "summary+window"}, nil
}

func (m *Manager) Summarize(ctx context.Context, messages []store.Message) (string, error) {
	var b strings.Builder
	for _, msg := range messages {
		fmt.Fprintf(&b, "[%s]: %s\n", msg.Role, msg.Content)
	}
	content := b.String()

	userPrompt := "请对以下对话历史生成简洁摘要，保留关键论点、立场和上下文：\n\n" + content

	resp, err := m.llmClient.Chat(ctx, []llm.ChatMessage{
		{Role: llm.RoleSystem, Content: summarizeSystemPrompt},
		{Role: llm.RoleUser, Content: userPrompt},
	}, &llm.ChatOptions{Temperature: 0.3})
	if err != nil {
		return "", fmt.Errorf("llm summarize: %w", err)
	}

	return resp.Content, nil
}

func (m *Manager) buildWindowMessages(systemPrompt string, messages []store.Message) []llm.ChatMessage {
	result := make([]llm.ChatMessage, 0, len(messages)+1)
	result = append(result, llm.ChatMessage{Role: llm.RoleSystem, Content: systemPrompt})

	for _, msg := range messages {
		var role llm.MessageRole
		switch msg.Role {
		case "user":
			role = llm.RoleUser
		case "assistant":
			role = llm.RoleAssistant
		default:
			continue
		}
		result = append(result, llm.ChatMessage{Role: role, Content: msg.Content})
	}

	return result
}
