# Dev Map

本文件记录项目开发导航图，服务于 `cc-propose`、`cc-apply`、`cc-review` 和新成员理解代码边界。
它只沉淀可复用、可验证、相对稳定的工程事实；临时猜测必须写入"待确认"，不得写成事实。

```text
last_updated: 2026-04-30
updated_by: cc-init
confidence: medium
```

## 1. 模块导航

| 模块 / 边界 | 主要路径 | 职责 | 入口 | 关键依赖 | 信心 |
|-------------|----------|------|------|----------|------|
| HTTP Server | `cmd/server/` | 路由、请求解析、响应、graceful shutdown、CORS 中间件 | `main.go` → `server.New()` → `server.Run()` | config, role, session, analysis | medium |
| 配置管理 | `internal/config/` | 多模型配置加载与校验 | `config.Load()` | 无 | medium |
| LLM 客户端 | `internal/llm/` | OpenAI 兼容协议适配 | `llm.Client` 接口 | 配置 | medium |
| 角色与目标 | `internal/role/` | 角色描述→目标推荐→维度映射 | `role.NewService()` + `role.NewHandler()` | LLM 客户端 | medium |
| 会话管理 | `internal/session/` | 创建/对话/结束/查询会话 | `session.NewService()` + `session.NewHandler()` | store, memory, LLM | medium |
| 对话记忆 | `internal/memory/` | 滑动窗口 + LLM 摘要压缩 | `memory.NewManager()` | LLM 客户端 | medium |
| 分析引擎 | `internal/analysis/` | 维度分析、报告生成、自动触发 | `analysis.NewEngine()` + `analysis.NewService()` + `analysis.NewHandler()` | LLM、role、session、store | medium |
| 数据持久化 | `internal/store/` | SQLite 读写、schema 初始化 | `store.Open()` + `store.NewSessionStore()` | 无 | medium |
| Web 前端 | `web/` | 设定页、对话页、报告页 SPA | `npm run dev` → Vite dev server | HTTP API | medium |

## 2. 关键链路

| 链路 | 入口 | 主要步骤 | 影响数据 | 常用验证 | 风险 |
|------|------|----------|----------|----------|------|
| 角色设定 → 目标推荐 | `POST /api/roles/recommend-goals` | 用户提交角色描述 → LLM 推荐目标 → 返回目标列表 | Role, TrainingGoal | `role/handler_test.go` + `role/service_test.go` | LLM 推荐质量 |
| 角色设定 → 维度映射 | `POST /api/roles/recommend-dimensions` | 用户提交角色+目标 → LLM 推荐维度 → 返回维度列表 | Role, Dimension | `role/handler_test.go` + `role/service_test.go` | LLM 推荐质量 |
| 对话主流程 | `POST /api/sessions/{id}/chat` | 角色注入 → LLM 生成回复 → 记忆更新 → 轮数检查 | Session, Message | `session/handler_test.go` + `session/service_test.go` | token 超限、API 失败 |
| 会话创建 | `POST /api/sessions` | 角色配置+目标+维度 → 创建会话 → 返回 session ID | Session | `session/handler_test.go` | 参数校验 |
| 会话结束 | `POST /api/sessions/{id}/end` | 状态校验 → 标记结束 → 返回摘要 | Session | `session/handler_test.go` | 状态流转 |
| 会话查询 | `GET /api/sessions/{id}` | 查询会话详情 | Session | `session/handler_test.go` | |
| 分析报告生成 | `POST /api/sessions/{id}/analyze` | 校验会话状态 → 构造分析 Prompt → LLM 分析 → JSON 解析容错 → MD 渲染 → 持久化 | AnalysisReport | `analysis/handler_test.go` + `analysis/service_test.go` + `analysis/engine_test.go` | 分析质量、JSON 解析失败 |
| 报告查询 | `GET /api/sessions/{id}/report` + `GET .../reports` | 查询最新/历史报告 | AnalysisReport | `analysis/handler_test.go` | |
| 自动触发分析 | 会话结束回调 | OnSessionEndFunc → analysisSvc.TriggerAnalysis | AnalysisReport | `session/service_test.go` + `analysis/service_test.go` | 回调遗漏、自动触发失败 |
| 模型配置切换 | 启动加载 | 配置文件 → Config 结构体 → LLM Client | LLMConfig | 启动测试 | 配置错误 |
| 前端路由 | 用户访问 | localhost:5173 → Vue Router → 设定/对话/报告页 → fetch API | SPA 页面 | 浏览器验证 | API 不可用 |

## 3. 测试与验证入口

| 范围 | 测试路径 / 命令 | 覆盖对象 | 适用场景 | 备注 |
|------|----------------|----------|----------|------|
| 全量构建 | `go build ./...` | 所有 Go 包 | L1 构建验证 | |
| 前端构建 | `cd web && npm run build` | Vue 3 + TypeScript | 前端类型检查 + 构建 | vue-tsc + vite build |
| 全量测试 | `go test ./...` | 所有 Go 包 | L2 回归 | |
| LLM 客户端 | `go test ./internal/llm/...` | LLM 适配层 | 单独调试 LLM 接入 | 可能需 mock LLM API |
| 角色模块 | `go test ./internal/role/...` | 角色推荐链路 | handler + service + model | |
| 会话模块 | `go test ./internal/session/...` | 会话生命周期 | handler + service | |
| 记忆管理 | `go test ./internal/memory/...` | 滑动窗口+摘要 | manager | |
| 持久化 | `go test ./internal/store/...` | session/message/analysis CRUD | SessionStore + AnalysisStore | |
| 分析引擎 | `go test ./internal/analysis/...` | Engine + Service + Handler | engine/service/handler | 95 tests total |
| Web 前端 | `cd web && npm run dev` | Vite dev server (localhost:5173) | 浏览器手工验证 | CORS 已配置允许 localhost:5173 |
| E2E 手工 | curl / 浏览器 | 全链路 | Phase 0 手工验证、Phase 1 浏览器验证 | |

## 4. 易错边界

| 边界 | 证据位置 | 影响 | 处理原则 | 信心 |
|------|----------|------|----------|------|
| LLM API Key 泄露 | `config.go`、`config.yaml` | 安全风险 | 只用环境变量或外部配置注入，不提交到仓库 | high |
| LLM 返回格式不稳定 | `role/service.go`、`session/service.go` | 目标/维度/分析推荐解析失败 | JSON 解析容错 + 重试 | medium |
| 长对话 token 超限 | `memory/manager.go` | 对话中断或质量下降 | 滑动窗口 + LLM 摘要压缩 | medium |
| 角色 Prompt 注入 | `role/service.go` | AI 角色设定被用户输入覆盖 | System Prompt 与 User Prompt 严格分离 | medium |
| 会话状态直接赋值 | `session/service.go`、`store/session.go` | 非法状态迁移 | 需集中校验入口 | medium |
| 跨模型兼容性 | `llm/client.go` | 切换模型后行为异常 | OpenAI 兼容协议标准化 + 多模型测试 | low |

## 5. Change 影响索引

| change_id | 影响模块 | 影响链路 | 关联验证 | 状态 |
|-----------|----------|----------|----------|------|
| analysis-engine | analysis, store, session, config, cmd/server | 分析报告生成、报告查询、自动触发 | 95 tests, V1-V12 | done |
| web-frontend | web, internal/server | 三页面 SPA、CORS 中间件 | 13 findings fixed, knowledge沉淀5条 | done |

## 6. 待确认事项

| 问题 | 影响范围 | 建议确认方式 | 优先级 |
|------|----------|--------------|--------|
| 会话状态迁移是否需集中校验 | session、store | 审查 `session/service.go` 状态变更路径 | P1 |
| Go HTTP 路由库选择 | HTTP Server | 评估 `net/http` vs `chi` vs `gin` | P2 |
| 项目仓库名称和位置 | 全项目 | 用户决定 | P1 |

## 更新规则

- `cc-init` 只写基础导航：模块候选、入口、测试入口和待确认事项。
- `cc-enrich-context` 负责补关键链路、易错边界、验证入口和信心等级。
- `cc-apply` 只有在代码结构、模块边界或验证入口发生实质变化时才更新本文件。
- 不得写入无证据的架构判断、敏感信息、临时调试日志或单次执行细节。
