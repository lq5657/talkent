# E2E Integration 模式

> 来源: `e2e-integration` change 归档

## 中间件顺序与 Goroutine 安全

- **关键规则**: `RecoveryMiddleware` 必须在 `TimeoutMiddleware` 的 **内层**（子 goroutine 内），不能在 Timeout 外层
- **原因**: `TimeoutMiddleware` 通过 `go func()` 在独立 goroutine 中执行下游 handler。如果 Recovery 在 Timeout 外层，`defer recover()` 在父 goroutine 中，无法捕获子 goroutine 的 panic，进程直接崩溃
- **正确顺序**: `RequestID → Timeout(goroutine: Recovery → CORS → Handler)`
- **错误顺序**: `RequestID → Recovery(父goroutine) → Timeout(goroutine: CORS → Handler)` — Recovery 捕获不到子 goroutine panic
- **验证**: 必须写集成回归测试，在子 goroutine 中 `panic("test")` 并断言返回 500，而非进程崩溃
- **适用范围**: 任何使用 goroutine 执行 handler 的超时/异步中间件模式

## LLM 日志脱敏

- **规则**: JSON 解析失败重试后仍失败时，禁止记录 LLM 原始输出内容
- **错误做法**: `"raw_output", truncated`（记录最多 500 字符 LLM 原始输出）
- **正确做法**: `"content_len", len(resp.Content)` + `"parse_error", parseErr.Error()`
- **依据**: `rules/security.md` 敏感信息处理规则
- **适用范围**: 所有 LLM 调用点的错误日志

## E2E 并发测试中的 LLM 限流

- **问题**: 并发创建多个会话时，LLM API 可能因限流返回错误，导致部分 session_id 为 null
- **应对**: curl 脚本中使用"先并发创建 + 逐个检查 null + sleep 1 + 单独重试"模式
- **适用范围**: 任何涉及 LLM API 并发调用的 E2E 测试脚本
