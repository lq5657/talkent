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

## 后端挂载前端静态文件的验证盲区

- **问题**: API 测试全部通过 + 前端 build 成功 ≠ 后端正确托管前端文件。服务端注册了 API 路由但忘记挂载 `web/dist/` 静态文件，`GET /` 返回 404
- **根因**: 这是"胶水代码"验证盲区 — API handler 有单元测试，前端代码有 `npm run build`，但**谁来把前端文件挂到 HTTP server 上**这个横切关注点没有被任务覆盖，也没有对应的验证步骤
- **验证缺口分析**:
  - `go test ./...` → 测试直接调 handler，不经过 ServeMux 路由匹配
  - E2E curl 脚本 → 全是 API 路径（`/api/*`），不访问 `/`
  - `npm run build` → 只验证前端编译，不验证产物被后端托管
  - `npm run dev` → 走 Vite dev server（:5173），不是 Go server（:8080）
- **预防措施**:
  1. 每个涉及前端 + 后端的 change，spec 和 tasks 中必须包含"静态文件挂载"这一横切关注点
  2. E2E 验证场景中至少包含一个 `curl -s http://localhost:8080/` 检查返回 200 和 HTML 内容
  3. 首次搭建或前后端联通变更时，必须手工浏览器验证首页可访问
- **适用范围**: 任何前后端分离但后端托管静态文件的项目
