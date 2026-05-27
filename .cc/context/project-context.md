---
alwaysApply: true
---
### 工程上下文

首次使用时执行 `cc-init` 让 AI 建立"基础事实层"并填充本文件。
当需要更完整的项目事实画像时，再执行 `cc-enrich-context` 补充"补充事实层"。

补充边界：
- `cc-init` 只负责识别后续命令高频复用、且能低成本确认的基础事实
- `cc-enrich-context` 负责补充高解释成本但仍属于项目事实的内容
- `cc-init` 与 `cc-enrich-context` 都只回写本文件，不负责创建脚手架资产
- 若要做存量项目体检，使用 `cc-inspect-codebase`
- 若要输出面向人的系统讲解材料，使用 `cc-explain-system`

#### 使用说明

- 本文件用于记录项目事实与事实边界，不用于强行套用预设架构
- 已有项目：优先记录"实际观察到"的目录、入口、依赖和约定
- 新项目：只记录当前已知事实与待确认事项，不在本文件中承载设计建议
- 所有无法低成本确认的内容统一写入"待确认事项"，不要把推测写成事实

#### 结构说明

本文件分为两层：

1. 基础事实层：由 `cc-init` 建立，长期驻留上下文，用于后续命令快速定位
2. 补充事实层：由 `cc-enrich-context` 补充，用于后续命令获得更完整但仍以事实为主的系统认知

领域语言另存于 `.cc/context/domain-language.md`。本文件可以引用领域语言事实，但不要把术语表复制到这里。

## 一、基础事实层（`cc-init` 填充）

#### 1. 项目基本身份（必填）

* 模式: 已有项目
* 当前阶段: 已接入规范
* 应用名: Talkent
* 简介: 角色扮演对话训练智能体 — 自由设定角色和场景，与 AI 进行 1v1 沉浸式对话训练，结束后获得多维度、结构化的表达分析反馈
* 运行形态: HTTP API + SPA + Android 原生客户端（多平台）

#### 2. 技术入口（必填）

* 构建工具: Go Modules (`go.mod`), npm/Vite (`web/package.json`), Gradle Kotlin DSL (`android/`)
* 依赖入口: `go.mod` (后端), `web/package.json` (前端), `android/build.gradle.kts` (Android)
* 启动入口: `cmd/server/main.go` (Go 服务), 命令: `go run ./cmd/server/` 或 `make run`; `android/` → `./gradlew assembleDebug` (Android APK)
* 多入口形态: 多平台（Go 服务 + Web SPA + Android 客户端）

#### 3. 关键目录导航（必填）

| 目录 | 已确认职责 | 备注 |
|------|------------|------|
| `cmd/server/` | 服务入口，组装依赖并启动 HTTP server | main.go 组装 config → db → llm → role/session/analysis → server |
| `internal/config/` | YAML 配置加载与环境变量覆盖 | 环境变量前缀 `TALKENT_` |
| `internal/llm/` | OpenAI 兼容 LLM 客户端 | 基于 go-openai |
| `internal/role/` | 角色设定 → 目标推荐 → 维度确定 | handler/service/model 三层 |
| `internal/session/` | 会话生命周期与对话管理 | handler/service/errors 三层，含轮数上限控制 |
| `internal/memory/` | 滑动窗口记忆与摘要压缩 | Manager 封装 |
| `internal/analysis/` | 多维度分析引擎与 Markdown 报告 | engine/service/handler 三层 |
| `internal/store/` | SQLite 持久化 | session/analysis 两个 store + db/schema |
| `internal/server/` | HTTP 路由、中间件与静态文件托管 | SPA fallback handler |
| `internal/log/` | slog 日志初始化 | 支持 file/stdout |
| `web/` | Vue 3 前端 SPA | Composition API + Tailwind CSS v4 + Vite |
| `android/` | Android 原生客户端 | Kotlin + Jetpack Compose + Room + Gradle Kotlin DSL |
| `test/e2e/` | E2E 验证场景与 curl 脚本 | 5 个场景 + run-all.sh |

#### 4. 配置与测试入口（必填）

| 项目 | 路径 / 命令 / 模式 | 状态与依据 |
|------|------------------|------------|
| 配置入口 | `config.yaml` → `internal/config/config.go` (YAML + 环境变量覆盖) | `confirmed` (config.go:51 Load函数) |
| 测试入口 / 目录 / 文件模式 | `go test ./...` / `go vet ./...`; 文件: `*_test.go` (13 个测试文件) | `confirmed` (每个 package 都有 _test.go) |
| 日志入口 | `internal/log/log.go` (基于 slog); 配置 `config.yaml` log.level/log.file | `confirmed` (log.go + config.go LogConfig) |

#### 5. 关键依赖（必填）

| 依赖/中间件 | 用途 | 备注 |
|-------------|------|------|
| `github.com/sashabaranov/go-openai` | OpenAI 兼容 API 客户端 | LLM 调用核心 |
| `modernc.org/sqlite` | 纯 Go SQLite 实现 | 无需 CGO |
| `gopkg.in/yaml.v3` | YAML 配置文件解析 | config.go 使用 |
| `github.com/google/uuid` | UUID 生成 | session/analysis ID + RequestID 中间件 |
| Vue 3 + vue-router | 前端框架与路由 | Composition API |
| Tailwind CSS v4 + Vite | 样式与构建 | Vite 开发服务器自动代理 API |
| marked + highlight.js + dompurify | Markdown 渲染与安全 | 分析报告展示 |

#### 6. 后续阅读导航（必填）

| 目标命令 | 建议优先阅读位置 | 备注 |
|----------|------------------|------|
| `cc-propose` | `internal/` 下对应模块代码 + `README.md` API 表 | 了解现有模块边界和 API 契约 |
| `cc-inspect-codebase` | `cmd/server/main.go` → `internal/` 逐包 | 追踪组装链路 |
| `cc-explain-system` | `README.md` 项目结构 + `cmd/server/main.go` | 快速理解整体架构 |
| `cc-test` | `go test ./...` + `test/e2e/curl/` | 已有测试覆盖 |
| 领域语言确认 | `.cc/context/domain-language.md` | 仅在涉及业务术语、状态名、产品概念或易混词时优先读取 |

## 二、补充事实层（`cc-enrich-context` 补充）

#### 7. 实际分层与调用关系（选填）

HTTP Handler (internal/server/middleware.go)
↓
Handler 层 (internal/{role,session,analysis}/handler.go)
↓
Service 层 (internal/{role,session,analysis}/service.go)
↓
Store 层 (internal/store/) + Engine (internal/analysis/engine.go) + LLM Client (internal/llm/)

**中间件栈顺序** (server.go:31-34):
RequestID → Timeout → Recovery → CORS → mux (从外到内)

| 中间件 | 职责 | 证据 |
|--------|------|------|
| RequestID | 生成 UUID 注入 context + X-Request-ID response header | middleware.go:17-23 |
| Timeout | goroutine + context.WithTimeout，超时返回 504 | middleware.go:56-82 |
| Recovery | panic 恢复，记录 request_id/method/path/stack，返回 500 | middleware.go:33-53 |
| CORS | 仅允许 localhost:5173 (Vite dev server) | server.go:42-51 |

#### 8. 代码约定与团队规范（选填）

| 主题 | 当前约定 | 证据 |
|------|----------|------|
| 日志 | slog TextHandler, AddSource: true, 级别: debug/warn/error/info | internal/log/log.go |
| 错误处理 | 包级 sentinel errors (errors.go) + fmt.Errorf wrap | session/errors.go, analysis/errors.go |
| 配置管理 | YAML 文件 + 环境变量覆盖 (TALKENT_ 前缀), 代码内置默认值 | config.go |
| 测试策略 | Go 标准 testing 包，每包有 _test.go (13 个); 前端无单元/组件测试 | find web/src -name "*test*" 返回空 |
| 并发模式 | database/sql 默认连接池; Timeout 中间件使用 goroutine + channel; 优雅关闭使用 signal channel | db.go, middleware.go, server.go:Run() |
| HTTP 响应 | JSON Content-Type, 统一 `{"error":"..."}` 格式 | middleware.go, server.go |

#### 9. 日志方案（选填）

| 项目 | 当前约定 | 备注 |
|------|----------|------|
| 输出目标 | stdout (默认) 或文件 (config.yaml log.file) | TextHandler, 非 JSON |
| 日志格式 | slog TextHandler + AddSource (文件名:行号) | 时间默认 slog 格式 |
| 切分策略 | 无（依赖外部 logrotate 或平台） | slog 本身不支持切分 |
| 保留策略 | 无 | |
| Trace/Request ID | UUID-based, 存储在 context ("request_id"), 响应头 X-Request-ID | middleware.go:17-23, RequestIDFromContext() |
| Panic 日志字段 | error, request_id, method, path, stack | middleware.go:39-44 |
| 关键业务日志 | session created/ended, chat round completed, auto analysis trigger | session/service.go |

#### 10. 配置管理（选填）

| 项目 | 当前约定 | 备注 |
|------|----------|------|
| 配置来源 | YAML 文件 + 环境变量 | config.go |
| 默认值策略 | 显式默认值（代码内置 Config 结构体） | config.go Load 函数 |
| 必填校验 | 启动期校验（config 文件读取失败 → os.Exit(1)） | |
| 敏感配置处理 | llm.api_key 建议通过 TALKENT_LLM_API_KEY 环境变量注入 | README + config.go |
| 环境差异 | 无多环境配置文件，单一 config.yaml | |
| Feature Flag | analysis.auto_trigger 配置开关 | config.go AnalysisConfig |
| 环境变量覆盖 | 全部 11 个 TALKENT_* 变量，包含 SERVER_HOST/PORT, DATABASE_PATH, LOG_LEVEL/FILE, LLM_PROVIDER/BASE_URL/API_KEY/MODEL, ANALYSIS_AUTO_TRIGGER, SESSION_MEMORY_WINDOW_SIZE | config.go:89-123 |

#### 11. 可观测性（选填）

| 项目 | 当前约定 | 备注 |
|------|----------|------|
| 关键日志字段 | request_id (通过 RequestIDFromContext 提取), session_id, error | middleware.go, session/service.go |
| Metrics | 无 | |
| Alerting | 无 | |
| Tracing | 无（仅 request_id 用于日志关联，非分布式 tracing） | |
| 健康检查 | GET /health → db.Ping() → {"status":"ok"} 或 503 | server.go:61-77 |

#### 12. 测试策略（选填）

| 项目 | 当前约定 | 备注 |
|------|----------|------|
| Repo 测试方式 | 真实 SQLite 文件 (temp file) | store 测试 |
| 链路回归方式 | `go test ./...` | 13 个 _test.go |
| 集成验证方式 | E2E curl 脚本 (5 场景, 全部 PASS 于 2026-04-30) | test/e2e/curl/run-all.sh |
| Bugfix 回归要求 | 待确认 | |
| 前端测试 | 无单元/组件测试文件 | find web/src 无 *_test.ts / *.spec.ts |
| 静态检查 | `go vet ./...` | |

#### 13. 外部依赖与集成边界（选填）

| 依赖类型 | 描述 | 备注 |
|----------|------|------|
| 数据库 | SQLite (modernc.org/sqlite, 纯 Go, 无 CGO) | talkent.db, 3 张表, 3 个索引 |
| 消息队列 | 无 | |
| RPC / HTTP 下游 | 无（纯服务端 + LLM API 调用） | |
| 第三方平台 | DeepSeek API (OpenAI 兼容) | 通过 go-openai 客户端 |

**数据库 Schema** (store/schema.go):
- `sessions`: id, role_config(JSON), goals(JSON), dimensions(JSON), status, round_limit, created_at, updated_at
- `messages`: id, session_id(FK), role(CHECK 'user'/'assistant'), content, sequence_num, created_at
- `analysis_reports`: id, session_id(FK), dimension_results(JSON), markdown_content, model_used, created_at
- 索引: messages(session_id), messages(session_id, sequence_num), analysis_reports(session_id)

**Migration 策略** (store/db.go:35-47):
- 启动时 ALTER TABLE 重跑，通过 "duplicate column" / "already exists" 错误抑制实现幂等
- 当前有 2 个 migration: analysis_reports 加 markdown_content 和 model_used 列

#### 13.1 关键链路索引（选填，最多 3-5 条）

| 优先级 | 链路 | 起点 | 关键中间层 | 终点/外部依赖 | 主要文件 | 备注 |
|--------|------|------|------------|---------------|----------|------|
| P0 | 对话训练全流程 | POST /api/sessions → POST /chat | session service → memory manager → LLM client | LLM API (DeepSeek) | internal/session/service.go, internal/memory/manager.go, internal/llm/ | 核心业务链路 |
| P0 | 分析报告生成 | POST /api/sessions/:id/analyze | analysis engine → LLM client → store | LLM API | internal/analysis/ | 分析结果持久化 |
| P0 | 自动分析触发 | Session.End → OnSessionEnd callback | session.notifySessionEnd → analysis.TriggerAnalysis | analysis engine | session/service.go:275-279, cmd/server/main.go:63-69 | context.Background() 异步触发 |
| P1 | 角色设定与推荐 | POST /api/roles/recommend-* | role service → LLM client | LLM API | internal/role/ | 会话前置步骤 |

#### 13.2 已知脆弱区域（选填）

| 区域 | 当前现象 | 影响命令 | 证据来源 | 备注 |
|------|----------|----------|----------|------|
| Migration 幂等 | ALTER TABLE 重跑 + 错误字符串匹配抑制 | cc-apply (DB 变更时) | store/db.go:35-47 | 无 migration 版本表，依赖错误消息文本 |
| OnSessionEnd context | 使用 context.Background() 而非请求 context | cc-test, 排障 | session/service.go:277 | 自动分析无超时/取消传播 |
| 轮数上限死循环 | 达轮数上限后重复 chat 曾导致死循环 | — | git 100dd81 | 已修复 |

#### 14. 领域特性与高风险点（选填）

| 类型 | 描述 | 备注 |
|------|------|------|
| 资金 | 无 | |
| 状态流转 | 会话状态: active → completed (仅两个状态) | 见下方状态机详述 |
| 权限 | 无（单用户本地应用） | |
| 外部依赖 | LLM API 调用（超时、失败处理）；摘要失败降级为 window-only | memory/manager.go:46-49 |

**会话状态机详述** (session/service.go + session/errors.go):

```
active ──┬── Chat() 达到 RoundLimit ──→ completed (auto)
         ├── EndSession() ────────────→ completed (manual)
         └── (无其他迁移路径)

已完成会话的 Chat() → ErrSessionCompleted
已完成会话的 EndSession() → 幂等返回已有结果 (非 error)
```

- 仅 2 个 sentinel error: `ErrSessionNotFound`, `ErrSessionCompleted`
- 没有 "analyzed" 状态 — 分析完成后 session 仍为 "completed"
- 手动结束和自动结束都调用 `notifySessionEnd()` → `OnSessionEnd` callback
- 轮数计算: `currentRound = msgCount / 2` (每轮 = user + assistant 两条消息)

**记忆窗口策略** (memory/manager.go:35-56):

```
消息总数 <= windowSize (默认10) → "window" 模式 (全部消息)
消息总数 > windowSize            → "summary+window" 模式
  ├── 超出部分 → LLM 摘要 (temperature=0.3)
  └── 窗口内部分 → 原文
摘要失败 → 降级为 "window" 模式 (仅窗口内消息 + warn log)
```

## 三、事实边界

#### 15. 已确认事实范围（必填）

- 已确认单仓库、主语言 (Go + TypeScript/Vue)、启动入口和配置入口
- 已确认测试入口、测试文件模式和 E2E 入口
- 已确认核心模块边界 (9 个 internal package + web 前端 + android/ 客户端) 和关键依赖
- 已确认项目运行形态和主要链路 (对话训练、分析报告、角色推荐、自动分析触发)
- **本次 enrich 新增确认**: 中间件栈顺序与职责、日志方案 (TextHandler + AddSource + RequestID)、会话状态机 (active→completed)、数据库 Schema (3 表 3 索引)、Migration 策略 (幂等重跑)、记忆窗口/摘要压缩策略 (窗口+摘要+降级)、前端测试情况 (无单元/组件测试)、健康检查端点、CORS 策略、优雅关闭 (5s 超时)、全部 11 个环境变量
- 尚未确认: 无 Metrics/Alerting/Tracing 方案（已确认"无"即为事实）、Bugfix 回归要求、环境差异策略（已确认"单一 config.yaml"即为事实）

#### 15.1 本轮确认依据（选填）

| 主题 | 主要依据 | 证据位置 | 备注 |
|------|----------|----------|------|
| 中间件栈与顺序 | server.go New() + middleware.go | internal/server/ | RequestID → Timeout → Recovery → CORS |
| 日志方案 | log.go (TextHandler+AddSource) + middleware.go (RequestID/Recovery 日志字段) | internal/log/, internal/server/ | |
| 会话状态机 | session/service.go Chat/EndSession + session/errors.go | internal/session/ | active→completed, 2 sentinel errors |
| 数据库 Schema | store/schema.go (DDL) + store/db.go (Open/migrations) | internal/store/ | 3 tables, 3 indexes, 2 migrations |
| 记忆窗口策略 | memory/manager.go BuildContext/Summarize | internal/memory/ | window/summary+window + 降级 |
| 前端测试情况 | find web/src -name "*test*" 返回空 | web/src/ | E2E curl 覆盖集成验证 |
| E2E 覆盖 | test/e2e/scenarios.md (5 场景全部 PASS) | test/e2e/ | 全链路/离线/空输入/轮数/并发 |

#### 16. 待确认事项（必填）

- Bugfix 回归要求待确认（是否要求每个 bugfix 至少一条回归证据）
- 环境差异策略已确认"无"（单一 config.yaml），如需多环境支持需新建
- Metrics/Alerting/Tracing 已确认"无"，如需添加需从零建设
