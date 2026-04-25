# Review — llm-client

## 审查信息

- **change_id**: llm-client
- **reviewed_at**: 2026-04-25
- **reviewed_by**: cc-review (main flow)

## Stage 1: Spec Compliance

**结论: PASSED**

| Spec 功能点 | 实现 | 合规 |
|------------|------|------|
| Client 接口定义，Chat(ctx, messages, opts) | `client.go:31-32` | ✅ |
| 领域类型 ChatMessage/ChatResponse/ChatOptions | `client.go:13-29`，不暴露 go-openai 类型 | ✅ |
| go-openai SDK 实现 | `openai.go:17-38` NewClient | ✅ |
| DefaultConfig + BaseURL 适配非 OpenAI 端点 | `openai.go:28-31` | ✅ |
| 单活跃 provider，复用 config.LLMConfig | 无多 provider 逻辑 | ✅ |
| cmd/server 集成 LLM 客户端初始化 | `main.go:37-43` | ✅ |
| 429/500 重试，最多 3 次，指数退避 | `openai.go:60-92` | ✅ |
| 401 不重试 | `openai.go:110-121` isRetryableError | ✅ |
| 日志不含 API Key | 仅记录 provider/model/messages/tokens | ✅ |
| context.Context 透传与取消响应 | `openai.go:70-74,77` | ✅ |
| V1-V6 验证映射闭环 | 19 tests passed, all apply-covered | ✅ |

## Stage 2: Code Quality

**风险镜头**: security, configuration, observability, source-driven-development

### 2.1 编码规范

- ✅ 命名、错误透传、日志规范均符合 `rules/coding-style.md`
- ✅ 无 panic、无吞错误、无 fmt.Println
- ✅ maxRetries 常量化

### 2.2 安全

- ✅ API Key 不硬编码，通过 config 注入
- ✅ 日志不含 API Key（测试 V5 验证）

### 2.3 可观测性

- ✅ 调用入口、结果、失败、重试均有结构化日志
- ✅ 关键字段：provider、model、messages、tokens、attempt、error

### 2.4 配置

- ✅ 复用 LLMConfig，不新增配置项
- ✅ nil config 校验

### 2.5 Source-Driven Development

- ✅ go-openai 使用方式与 Context7 文档一致
- ✅ `DefaultConfig` + `BaseURL` 适配方式已验证

## Findings

### F1 — Minor: NewClient 不校验 logger 为 nil

- **位置**: `openai.go:23`
- **描述**: `NewClient(cfg, nil)` 不会报错，但 `Chat()` 会 panic
- **风险等级**: Minor
- **状态**: accepted
- **接受理由**: logger 由 main.go 统一初始化，实际不可能为 nil；后续加固时处理

### F2 — Minor: NewClient 不校验 APIKey 为空

- **位置**: `openai.go:23`
- **描述**: 空 API Key 不会在初始化阶段报错，调用时才返回 401
- **风险等级**: Minor
- **状态**: accepted
- **接受理由**: 空 Key 调用时 LLM 返回 401，错误信息可定位；增加校验是优化项

### F3 — Minor: 重试耗尽错误不包含重试历史

- **位置**: `openai.go:94-96`
- **描述**: 仅返回最后一次 err，不包含"连续 N 次失败"上下文
- **风险等级**: Minor
- **状态**: accepted
- **接受理由**: 日志已记录每次重试，排障信息充足

### F4 — Important: config.LLMConfig.Timeout 未应用到 HTTP 客户端

- **位置**: `openai.go:28-31`
- **描述**: `NewClient` 使用 `openai.DefaultConfig(cfg.APIKey)` 但未将 `cfg.Timeout` 设置到 HTTP client。spec §8 声明"超时由 config.LLMConfig.Timeout 控制"，但实现未兑现。如果调用方忘记设置 context 超时，LLM 调用无超时保护，可能导致 goroutine 泄漏。
- **风险等级**: Important
- **状态**: fixed
- **修复内容**: 在 `NewClient` 中新增 `if cfg.Timeout > 0 { ocfg.HTTPClient = &http.Client{Timeout: cfg.Timeout} }`；新增 `TestNewClient_ConfigTimeoutApplied` 和 `TestChatCompletion_HTTPTimesOutWithConfigTimeout` 测试
- **修复验证**: 21 tests passed, go build + vet PASSED

## Task Coverage Matrix

| Task | 覆盖文件 | 验证证据 | 状态 |
|------|----------|----------|------|
| Task 1 | client.go, go.mod | go build PASSED | done |
| Task 2+3 | openai.go, convert.go, main.go | go build + vet PASSED | done |
| Task 4 | openai_test.go | 21 tests passed (含 F4 修复新增 2 tests) | done |

## 验证映射状态

| ID | 需求项 | 最低等级 | 证据 | 闭环 |
|----|--------|---------|------|------|
| V1 | Client 接口可编译 | L2 | go build | apply-covered |
| V2 | ChatCompletion 参数传递 | L2 | TestChatCompletion_Success | apply-covered |
| V3 | 自定义 BaseURL | L2 | TestChatCompletion_CustomBaseURL | apply-covered |
| V4 | 错误分类处理 | L2 | TestRetryOn429/500, NoRetryOn401, RetryExhausted | apply-covered |
| V5 | API Key 不在日志 | L2 | TestChatCompletion_APIKeyNotInLogs | apply-covered |
| V6 | context 超时控制 | L2 | TestChatCompletion_ContextCancellation + TestChatCompletion_HTTPTimesOutWithConfigTimeout | apply-covered |

## 总体结论

- Stage 1: **PASSED**
- Stage 2: **PASSED**
- F4 (Important): **fixed**，有 fresh evidence
- F1-F3 (Minor): accepted
- **审查结论**: 可归档
