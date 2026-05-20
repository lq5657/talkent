package llm

import "context"

type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

type ChatMessage struct {
	Role    MessageRole
	Content string
}

type ChatResponse struct {
	Content          string
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type ChatOptions struct {
	Temperature float32
	MaxTokens   int
}

type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

type Client interface {
	Chat(ctx context.Context, messages []ChatMessage, opts *ChatOptions) (*ChatResponse, error)
	ChatStream(ctx context.Context, messages []ChatMessage, opts *ChatOptions) (<-chan StreamChunk, error)
}
