package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/lq5657/talkent/internal/llm"
	"github.com/lq5657/talkent/internal/role"
	"github.com/lq5657/talkent/internal/store"
)

type DimensionResult struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Score       int      `json:"score"`
	Comment     string   `json:"comment"`
	Suggestions []string `json:"suggestions"`
}

type AnalysisResult struct {
	DimensionResults []DimensionResult
	Markdown         string
	ModelUsed        string
}

type Engine struct {
	llmClient llm.Client
	logger    *slog.Logger
}

func NewEngine(llmClient llm.Client, logger *slog.Logger) *Engine {
	return &Engine{
		llmClient: llmClient,
		logger:    logger,
	}
}

func (e *Engine) Analyze(ctx context.Context, roleDesc, scenario string, messages []store.Message, dimensions []role.Dimension) (*AnalysisResult, error) {
	prompt := e.buildPrompt(roleDesc, scenario, messages, dimensions)

	resp, err := e.callWithRetry(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("llm analysis: %w", err)
	}

	e.logger.Info("analysis completed", "model", resp.Model, "dimension_count", len(dimensions))

	results, err := e.parseDimensionResults(resp.Content)
	if err != nil {
		return nil, err
	}

	markdown := e.renderMarkdown(roleDesc, scenario, results)

	return &AnalysisResult{
		DimensionResults: results,
		Markdown:         markdown,
		ModelUsed:        resp.Model,
	}, nil
}

func (e *Engine) buildPrompt(roleDesc, scenario string, messages []store.Message, dimensions []role.Dimension) []llm.ChatMessage {
	var dimDescs []string
	for _, d := range dimensions {
		dimDescs = append(dimDescs, fmt.Sprintf("- %s: %s", d.Name, d.Description))
	}

	var dialogue strings.Builder
	for _, m := range messages {
		label := "用户"
		if m.Role == "assistant" {
			label = "AI训练师"
		}
		fmt.Fprintf(&dialogue, "[%s]: %s\n", label, m.Content)
	}

	systemPrompt := `你是一位专业的对话分析专家。你需要根据给定的分析维度，对对话内容进行深度分析。

你必须以纯 JSON 格式输出分析结果，不要输出任何其他内容。JSON 格式如下：

{
  "dimensions": [
    {
      "name": "维度名称",
      "description": "维度描述",
      "score": 8,
      "comment": "评语",
      "suggestions": ["建议1", "建议2"]
    }
  ]
}

要求：
- score 为 1-10 的整数
- comment 为 50-150 字的评语
- suggestions 为 1-3 条具体可操作的建议
- 只输出 JSON，不要输出其他文字`

	var userPrompt strings.Builder
	fmt.Fprintf(&userPrompt, "## 角色设定\n%s\n\n", roleDesc)
	if scenario != "" {
		fmt.Fprintf(&userPrompt, "## 场景\n%s\n\n", scenario)
	}
	fmt.Fprintf(&userPrompt, "## 分析维度\n%s\n\n", strings.Join(dimDescs, "\n"))
	fmt.Fprintf(&userPrompt, "## 对话内容\n%s\n\n", dialogue.String())
	userPrompt.WriteString("请严格按上述 JSON 格式输出分析结果。")

	return []llm.ChatMessage{
		{Role: llm.RoleSystem, Content: systemPrompt},
		{Role: llm.RoleUser, Content: userPrompt.String()},
	}
}

type llmAnalysisResponse struct {
	Dimensions []DimensionResult `json:"dimensions"`
}

func (e *Engine) parseDimensionResults(content string) ([]DimensionResult, error) {
	content = strings.TrimSpace(content)

	// Strip markdown code fences if present
	if strings.HasPrefix(content, "```") {
		if idx := strings.Index(content[3:], "\n"); idx >= 0 {
			content = content[3+idx+1:]
		}
		if idx := strings.LastIndex(content, "```"); idx >= 0 {
			content = content[:idx]
		}
		content = strings.TrimSpace(content)
	}

	var resp llmAnalysisResponse
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		return nil, fmt.Errorf("parse analysis json: %w", err)
	}

	if len(resp.Dimensions) == 0 {
		return nil, fmt.Errorf("analysis returned no dimensions")
	}

	return resp.Dimensions, nil
}

func (e *Engine) callWithRetry(ctx context.Context, messages []llm.ChatMessage) (*llm.ChatResponse, error) {
	resp, err := e.llmClient.Chat(ctx, messages, &llm.ChatOptions{Temperature: 0.3})
	if err != nil {
		return nil, err
	}

	// Try parsing; on failure, retry once with stronger format instruction
	if _, parseErr := e.parseDimensionResults(resp.Content); parseErr != nil {
		e.logger.Warn("analysis json parse failed, retrying", "error", parseErr)

		retryMsgs := append(messages, llm.ChatMessage{
			Role:    llm.RoleAssistant,
			Content: resp.Content,
		}, llm.ChatMessage{
			Role:    llm.RoleUser,
			Content: "你上次的输出不是合法的 JSON 格式。请严格按要求的 JSON 格式重新输出，只输出 JSON，不要有任何其他文字。",
		})

		resp, err = e.llmClient.Chat(ctx, retryMsgs, &llm.ChatOptions{Temperature: 0.1})
		if err != nil {
			return nil, err
		}

		if _, parseErr = e.parseDimensionResults(resp.Content); parseErr != nil {
			truncated := resp.Content
			if len(truncated) > 500 {
				truncated = truncated[:500] + "..."
			}
			e.logger.Error("analysis json parse failed after retry", "raw_output", truncated)
			return nil, fmt.Errorf("analysis json parse failed after retry: %w", parseErr)
		}
	}

	return resp, nil
}

func (e *Engine) renderMarkdown(roleDesc, scenario string, results []DimensionResult) string {
	var b strings.Builder

	b.WriteString("# 对话分析报告\n\n")

	b.WriteString("## 会话信息\n\n")
	fmt.Fprintf(&b, "- **角色设定**: %s\n", roleDesc)
	if scenario != "" {
		fmt.Fprintf(&b, "- **场景**: %s\n", scenario)
	}
	fmt.Fprintf(&b, "- **分析维度数**: %d\n\n", len(results))

	b.WriteString("## 维度评分\n\n")
	for _, d := range results {
		fmt.Fprintf(&b, "### %s（%d/10）\n\n", d.Name, d.Score)
		fmt.Fprintf(&b, "**描述**: %s\n\n", d.Description)
		fmt.Fprintf(&b, "**评语**: %s\n\n", d.Comment)
		if len(d.Suggestions) > 0 {
			b.WriteString("**改进建议**:\n")
			for _, s := range d.Suggestions {
				fmt.Fprintf(&b, "- %s\n", s)
			}
			b.WriteString("\n")
		}
	}

	var totalScore int
	for _, d := range results {
		totalScore += d.Score
	}
	avg := float64(totalScore) / float64(len(results))
	fmt.Fprintf(&b, "## 综合评分\n\n**%.1f / 10**\n", avg)

	return b.String()
}
