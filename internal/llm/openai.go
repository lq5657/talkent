package llm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/lq5657/talkent/internal/config"
	openai "github.com/sashabaranov/go-openai"
)

const maxRetries = 3

type OpenAIClient struct {
	client *openai.Client
	cfg    *config.LLMConfig
	logger *slog.Logger
}

func NewClient(cfg *config.LLMConfig, logger *slog.Logger) (*OpenAIClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("llm config is required")
	}

	ocfg := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		ocfg.BaseURL = cfg.BaseURL
	}

	return &OpenAIClient{
		client: openai.NewClientWithConfig(ocfg),
		cfg:    cfg,
		logger: logger,
	}, nil
}

func (c *OpenAIClient) Chat(ctx context.Context, messages []ChatMessage, opts *ChatOptions) (*ChatResponse, error) {
	req := openai.ChatCompletionRequest{
		Model:    c.cfg.Model,
		Messages: toOpenAIMessages(messages),
	}

	if opts != nil {
		req.Temperature = opts.Temperature
		req.MaxTokens = opts.MaxTokens
	}

	c.logger.Info("llm chat request",
		"provider", c.cfg.Provider,
		"model", c.cfg.Model,
		"messages", len(messages),
	)

	var resp openai.ChatCompletionResponse
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
			c.logger.Info("llm retry",
				"provider", c.cfg.Provider,
				"model", c.cfg.Model,
				"attempt", attempt,
				"backoff", backoff,
			)

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, fmt.Errorf("llm chat cancelled during retry: %w", ctx.Err())
			}
		}

		resp, err = c.client.CreateChatCompletion(ctx, req)
		if err == nil {
			break
		}

		if !isRetryableError(err) {
			break
		}

		c.logger.Error("llm chat error",
			"provider", c.cfg.Provider,
			"model", c.cfg.Model,
			"attempt", attempt,
			"error", err,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("llm chat failed: %w", err)
	}

	result := fromChatResponse(resp)
	c.logger.Info("llm chat response",
		"provider", c.cfg.Provider,
		"model", result.Model,
		"prompt_tokens", result.PromptTokens,
		"completion_tokens", result.CompletionTokens,
		"total_tokens", result.TotalTokens,
	)

	return result, nil
}

func isRetryableError(err error) bool {
	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.HTTPStatusCode {
		case 429, 500, 502, 503:
			return true
		default:
			return false
		}
	}
	return false
}
