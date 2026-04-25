package llm

import openai "github.com/sashabaranov/go-openai"

func toOpenAIMessages(messages []ChatMessage) []openai.ChatCompletionMessage {
	result := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		result[i] = openai.ChatCompletionMessage{
			Role:    string(m.Role),
			Content: m.Content,
		}
	}
	return result
}

func fromChatResponse(resp openai.ChatCompletionResponse) *ChatResponse {
	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}
	return &ChatResponse{
		Content:          content,
		Model:            resp.Model,
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}
}
