# LLM Client 集成模式

> 来源: `llm-client` change 归档

## go-openai SDK 集成

- 使用 `openai.DefaultConfig(apiKey)` + `config.BaseURL` 适配非 OpenAI 端点（DeepSeek、通义千问等）
- 创建自定义 HTTP client 并赋值到 `ocfg.HTTPClient` 以支持超时配置
- `openai.APIError` 类型支持按 HTTP 状态码分类处理（429/500/502/503 可重试）

## 重试策略

- 指数退避：`2^(attempt-1)` 秒，最多 3 次重试
- 可重试：429、500、502、503
- 不可重试：401、403、400、404 等
- 退避期间检查 context 取消，避免无谓等待

## 领域类型解耦

- 定义 `ChatMessage`/`ChatResponse`/`ChatOptions` 领域类型，不暴露 go-openai SDK 类型
- `Client` 接口只有一个 `Chat` 方法，便于 mock 和替换实现
- `toOpenAIMessages` / `fromChatResponse` 转换函数集中处理映射

## 测试模式

- `httptest.NewServer` 模拟 LLM API 端点，零外部依赖
- 按状态码分场景测试（Success/429/500/401/503耗尽/context取消/日志脱敏）
- 用 `strings.Builder` 捕获 slog 输出验证日志脱敏

## 结构化 JSON 输出容错模式

> 来源: `analysis-engine` change

- System Prompt 明确声明 JSON 格式要求和 Schema，User Prompt 末尾追加格式提醒强化约束
- `parseDimensionResults` 先 `strings.TrimSpace`，再脱去 markdown code fence（`` ```json ... ``` ``），最后 `json.Unmarshal`
- `callWithRetry` 模式：首次调用 temperature=0.3，解析失败追加 assistant+user 消息后 temperature=0.1 重试一次
- 重试仍失败：截断原始输出至 500 字符记 ERROR 日志，返回错误（不无限重试）
- 成功路径：`callWithRetry` 直接返回解析结果，避免上层重复解析
- 测试模式：`mockLLMClient.responses []string` 按调用顺序返回不同响应，`callCount` 验证重试次数
