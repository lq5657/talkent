# Dev Map

本文件记录项目开发导航图，服务于 `cc-propose`、`cc-apply`、`cc-review` 和新成员理解代码边界。
它只沉淀可复用、可验证、相对稳定的工程事实；临时猜测必须写入"待确认"，不得写成事实。

```text
last_updated: 2026-05-24
updated_by: cc-init
confidence: high
```

## 1. 模块导航

| 模块 / 边界 | 主要路径 | 职责 | 入口 | 关键依赖 | 信心 |
|-------------|----------|------|------|----------|------|
| 服务入口 | `cmd/server/` | 依赖组装 + 启动 HTTP server + 注册 OnSessionEnd callback | `main.go` | 所有 internal 包 | high |
| 配置 | `internal/config/` | YAML 加载 + 11 个 TALKENT_* 环境变量覆盖 | `config.go:Load()` | yaml.v3 | high |
| LLM 客户端 | `internal/llm/` | OpenAI 兼容 API 封装 (Chat/ChatOptions/ChatStream) + SSE 流式响应 | `client.go`, `openai.go`, `convert.go` | go-openai | high |
| 角色管理 | `internal/role/` | 角色设定 → LLM 推荐目标 → LLM 推荐维度 | `handler.go`, `service.go`, `model.go`, `template.go` | LLM 客户端 | high |
| 会话管理 | `internal/session/` | 会话生命周期 (active→completed) + 对话 (Chat) + 轮数控制 + 系统提示词构建 | `handler.go`, `service.go`, `errors.go` | Store, Memory, LLM, role | high |
| 记忆管理 | `internal/memory/` | 滑动窗口记忆 + LLM 摘要压缩 + 降级策略 | `manager.go` | LLM 客户端 | high |
| 分析引擎 | `internal/analysis/` | 多维度 LLM 分析 + Markdown 报告生成 | `engine.go`, `service.go`, `handler.go`, `errors.go` | LLM, Store | high |
| 持久化 | `internal/store/` | SQLite CRUD + Schema DDL + 幂等 Migration | `db.go`, `schema.go`, `session.go`, `analysis.go` | modernc.org/sqlite | high |
| HTTP 服务 | `internal/server/` | 路由注册、4 层中间件栈、静态文件、SPA fallback、优雅关闭 | `server.go`, `middleware.go` | 所有 handler | high |
| 日志 | `internal/log/` | slog TextHandler 初始化 (AddSource=true) | `log.go` | log/slog | high |
| 前端 SPA | `web/` | Vue 3 + Tailwind CSS v4 + Vite, 3 个视图 (Setup/Chat/Report) | `src/main.ts` → `App.vue` → `router/` → `views/` | vue-router, marked, dompurify, highlight.js | high |
| E2E 验证 | `test/e2e/` | 5 个 curl 场景 + run-all.sh (全部 PASS 于 2026-04-30) | `curl/run-all.sh` | 后端 API | high |

## 2. 关键链路

| 链路 | 入口 | 主要步骤 | 影响数据 | 常用验证 | 风险 |
|------|------|----------|----------|----------|------|
| 对话训练全流程 | POST /api/sessions → POST /chat ×N → POST /end → POST /analyze → GET /report | 创建会话(active) → 多轮对话(LLM+记忆窗口) → 结束(completed) → 分析(LLM) → 报告 | sessions, messages, analysis_reports | E2E scenario-1-full-flow.sh | LLM API 超时/限流; 摘要压缩失败降级 |
| 自动分析触发 | Session.End → notifySessionEnd → OnSessionEnd callback | 会话结束 → context.Background() 异步 → analysis.TriggerAnalysis("auto") | analysis_reports | 需集成验证 | 异步失败静默 (仅 warn log); 无超时传播 |
| 角色设定与推荐 | POST /api/roles/recommend-goals → POST /api/roles/recommend-dimensions | 输入角色描述 → LLM 推荐目标 → LLM 推荐维度 | 无持久化（纯 LLM 调用） | role handler/service tests | LLM 返回 JSON 格式不稳定 |
| 轮数上限自动结束 | Chat() 检测 currentRound >= RoundLimit | 持久化 assistant 消息 → UpdateSessionStatus("completed") → notifySessionEnd | sessions.status, messages | E2E scenario-4-round-limit.sh | 已修复死循环 (commit 100dd81) |
| 摘要压缩降级 | memory.BuildContext() → Summarize() 失败 | LLM 摘要调用失败 → warn log → 降级为 window-only | LLM 上下文质量 | memory/manager_test.go | 降级后上下文可能丢失早期对话关键信息 |

## 3. 测试与验证入口

| 范围 | 测试路径 / 命令 | 覆盖对象 | 适用场景 | 备注 |
|------|----------------|----------|----------|------|
| 全量单元测试 | `go test ./...` | 所有 Go package (13 个 _test.go) | 每次改动后 | |
| 静态检查 | `go vet ./...` | 所有 Go package | PR 前 | |
| 全量 E2E | `./test/e2e/curl/run-all.sh` | 5 个核心场景 | 部署前 / 重大改动 | 需服务运行在 localhost:8080 |
| 单场景 E2E | `BASE_URL=http://localhost:8080 ./test/e2e/curl/run-all.sh N` | 单个场景 | 针对性验证 | N=1..5 |
| E2E 场景 1 | `test/e2e/curl/scenario-1-full-flow.sh` | 完整流程 (角色→会话→对话→结束→分析→报告) | 核心功能回归 | 7 步全链路 |
| E2E 场景 2 | 手工浏览器验证 | 离线 banner + 重试 | 前端网络不可达 | 需手动操作浏览器 |
| E2E 场景 3 | `test/e2e/curl/scenario-3-empty-input.sh` | 空角色描述/空消息/不存在会话/活跃会话分析 | 输入校验 + 错误码 | 4 项全部 4xx |
| E2E 场景 4 | `test/e2e/curl/scenario-4-round-limit.sh` | round_limit=1 + 超限拒绝 | 会话自动结束 | is_last + 409 |
| E2E 场景 5 | `test/e2e/curl/scenario-5-concurrent.sh` | 3 并发创建会话 + 消息隔离 | 并发安全 | 含 LLM 限流重试 |
| 前端开发 | `cd web && npm run dev` | 前端 SPA (Vite :5173, 代理 API :8080) | 前端开发调试 | CORS 仅允许 :5173 |
| 前端构建 | `cd web && npm run build` | 生产构建 → web/dist/ | 部署前 | 后端托管静态文件 |
| 健康检查 | `curl localhost:8080/health` | DB 连通性 | 部署后 | {"status":"ok"} 或 503 |

## 4. 易错边界

| 边界 | 证据位置 | 影响 | 处理原则 | 信心 |
|------|----------|------|----------|------|
| 会话状态流转 | `internal/session/service.go:115-116` (Chat), `:203` (EndSession) | 已完成会话的 chat/end 操作返回 ErrSessionCompleted | 必须通过 session service 校验，禁止直接修改 status 字段 | high |
| LLM 超时 | `internal/config/config.go:65` (LLMConfig.Timeout: 30s) | 对话/分析可能因 LLM 超时失败 | 通过 go-openai 客户端超时控制，错误向上传播 | high |
| SQLite 并发 | `internal/store/db.go:12` (sql.Open, 默认连接池) | 单文件 SQLite 写并发限制 | database/sql 默认连接池（未显式调优）；并发测试已通过 (E2E scenario-5) | medium |
| 前端安全 | `web/src/components/MarkdownRenderer.vue` (dompurify) | Markdown 内容可能含 XSS | 渲染前通过 dompurify 净化 | high |
| 记忆窗口降级 | `internal/memory/manager.go:46-49` | LLM 摘要失败 → window-only，丢失早期上下文 | 仅 warn log，不阻断对话 | high |
| 轮数上限死循环 | git log 100dd81 | 达轮数上限后 chat 进入死循环 | 已修复: Chat() 在 isLast 时先 UpdateSessionStatus 再返回 | high |
| Migration 幂等 | `internal/store/db.go:35-47` | ALTER TABLE 重跑依赖错误字符串匹配 | 跨 SQLite 版本/驱动版本的错误消息可能变化 | medium |
| OnSessionEnd context | `internal/session/service.go:277` (context.Background()) | 自动分析无请求超时/取消传播 | 仅适用于 fire-and-forget 场景；长时间分析可能无超时保护 | medium |
| 环境变量解析 | `internal/config/config.go:89-123` (fmt.Sscanf) | PORT 等数值字段解析无错误处理 | 非法值静默为 0，可能产生非预期行为 | low |

## 5. Change 影响索引

| change_id | 影响模块 | 影响链路 | 关联验证 | 状态 |
|-----------|----------|----------|----------|------|
| message-timing | session handler/service, ChatView, MessageBubble | Chat API 响应时序 + 消息时间展示 | V1 (L2 package), V2 (L4 manual) | done |
| voice-interaction | llm, session, server, ChatView, ChatInput, MessageBubble | SSE ChatStream + 浏览器语音交互 (STT/TTS) | L4 manual verified, 5 findings fixed | done |

## 6. 待确认事项

| 问题 | 影响范围 | 建议确认方式 | 优先级 |
|------|----------|--------------|--------|
| database/sql 连接池参数调优 | store 层并发性能 | 确认是否需要显式设置 MaxOpenConns | P3 |
| Bugfix 回归测试要求 | 测试策略 | 确认团队是否要求每个 bugfix 至少一条回归证据 | P2 |

## 更新规则

- `cc-init` 只写基础导航：模块候选、入口、测试入口和待确认事项。
- `cc-enrich-context` 负责补关键链路、易错边界、验证入口和信心等级。
- `cc-apply` 只有在代码结构、模块边界或验证入口发生实质变化时才更新本文件。
- 不得写入无证据的架构判断、敏感信息、临时调试日志或单次执行细节。
