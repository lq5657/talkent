---
change_id: analysis-engine
created: 2026-04-27
updated: 2026-04-27
---

### 任务拆分 — 分析引擎

#### 前置条件

* [ ] `spec.md` 已确认且 `status = apply`
* [ ] HARD-GATE 已通过
* [ ] `scaffold-project` 已完成
* [ ] `llm-client` 已完成
* [ ] `role-and-goal` 已完成
* [ ] `chat-session` 已完成

#### Task 1: Analysis Store CRUD + Schema 迁移

* **目标** : 实现 AnalysisReport 的数据库读写操作 + Schema 迁移新增 2 列
* **不包含范围** : 不实现 Engine（属于 T2），不实现 Service（属于 T3），不编写测试
* **涉及文件** :
  * `internal/store/analysis.go` — 新建，AnalysisReport 结构体 + CRUD
  * `internal/store/schema.go` — 修改，DDL 新增 markdown_content 和 model_used 列
* **关键签名** :
  ```go
  type AnalysisReport struct {
      ID               int64
      SessionID        string
      DimensionResults string // JSON
      MarkdownContent  string
      ModelUsed        string
      CreatedAt        time.Time
  }

  type AnalysisStore struct {
      db *sql.DB
  }

  func NewAnalysisStore(db *sql.DB) *AnalysisStore
  func (s *AnalysisStore) CreateReport(ctx context.Context, report *AnalysisReport) error
  func (s *AnalysisStore) GetLatestReport(ctx context.Context, sessionID string) (*AnalysisReport, error)
  func (s *AnalysisStore) ListReports(ctx context.Context, sessionID string) ([]AnalysisReport, error)
  ```
* **验收标准** : `go build ./...` 通过；AnalysisReport CRUD 方法可正常读写；Schema 迁移后 markdown_content 和 model_used 列存在且可读写
* **验证步骤** : `go build ./...`（V1/V11 编译级，完整功能验证在测试 task）
* **测试要求** : L2，V1/V11 在后续测试闭环
* **mapping_ids** : [V1, V11]
* **依赖 / Wave** : wave-1，无前置依赖
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/pre-apply.json -> baseline/post-task-1.json`

#### Task 2: Analysis Engine

* **目标** : 实现分析 Prompt 构造、LLM 调用、JSON 解析容错（一次重试）和 Markdown 报告渲染
* **不包含范围** : 不实现 Service（属于 T3），不实现 Handler（属于 T4），不编写测试
* **涉及文件** :
  * `internal/analysis/engine.go` — 新建，Engine 结构体 + Analyze + renderMarkdown
* **关键签名** :
  ```go
  type DimensionResult struct {
      Name        string `json:"name"`
      Description string `json:"description"`
      Score       int    `json:"score"`
      Comment     string `json:"comment"`
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

  func NewEngine(llmClient llm.Client, logger *slog.Logger) *Engine
  func (e *Engine) Analyze(ctx context.Context, roleDesc, scenario string, messages []store.Message, dimensions []role.Dimension) (*AnalysisResult, error)
  ```
* **验收标准** : `go build ./...` 通过；Analyze 构造的分析 Prompt 包含角色设定+场景+对话原文+维度列表+JSON 格式要求；JSON 解析失败自动重试一次；重试仍失败返回错误并记录原始输出截断至 500 字符；Markdown 报告包含标题+会话信息+维度评分/评语/建议
* **验证步骤** : `go build ./...`（V2/V3/V4/V5/V6 编译级，完整功能验证在测试 task）
* **测试要求** : L2，V2/V3/V4/V5/V6 在后续测试闭环
* **mapping_ids** : [V2, V3, V4, V5, V6]
* **依赖 / Wave** : wave-2，依赖 T1（AnalysisReport 结构体）
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-task-1.json -> baseline/post-task-2.json`

#### Task 3: Analysis Service + 自动触发钩子

* **目标** : 实现分析生命周期编排 + session.Service 的 OnSessionEndFunc 回调支持
* **不包含范围** : 不实现 Handler（属于 T4），不编写测试
* **涉及文件** :
  * `internal/analysis/service.go` — 新建，Service 结构体 + TriggerAnalysis/GetLatestReport/ListReports
  * `internal/analysis/errors.go` — 新建，哨兵错误
  * `internal/session/service.go` — 修改，新增 OnSessionEndFunc 字段和调用点
* **关键签名** :
  ```go
  // internal/analysis/service.go
  type Service struct {
      store  *store.AnalysisStore
      engine *Engine
      sessionStore *store.SessionStore
      logger *slog.Logger
  }

  func NewService(store *store.AnalysisStore, engine *Engine, sessionStore *store.SessionStore, logger *slog.Logger) *Service
  func (s *Service) TriggerAnalysis(ctx context.Context, sessionID string) (*AnalysisResult, error)
  func (s *Service) GetLatestReport(ctx context.Context, sessionID string) (*store.AnalysisReport, error)
  func (s *Service) ListReports(ctx context.Context, sessionID string) ([]store.AnalysisReport, error)

  // internal/session/service.go 新增
  type OnSessionEndFunc func(ctx context.Context, sessionID string)

  // Service 新增字段
  OnSessionEnd OnSessionEndFunc
  ```
* **验收标准** : `go build ./...` 通过；TriggerAnalysis 校验会话状态为 completed；调用 Engine.Analyze 并持久化；GetLatestReport/ListReports 正常工作；EndSession 成功后调用 OnSessionEnd 回调；自动触发失败不影响会话结束
* **验证步骤** : `go build ./...`（V7/V9 编译级，完整功能验证在测试 task）
* **测试要求** : L2，V7/V9 在后续测试闭环
* **mapping_ids** : [V7, V9]
* **依赖 / Wave** : wave-3，依赖 T1、T2
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-task-2.json -> baseline/post-task-3.json`

#### Task 4: HTTP Handler + 配置 + main.go 接线 + 测试

* **目标** : 实现 3 个 HTTP 端点 + 配置 + main.go 接线 + 全部单元测试
* **不包含范围** : 不测试真实 LLM API
* **涉及文件** :
  * `internal/analysis/handler.go` — 新建，HTTP Handler + 3 个端点
  * `internal/config/config.go` — 修改，新增 AnalysisConfig + 环境变量覆盖
  * `cmd/server/main.go` — 修改，接线 analysis 模块 + 自动触发钩子
  * `internal/store/analysis_test.go` — 新建，Store CRUD 测试
  * `internal/analysis/engine_test.go` — 新建，Engine 测试
  * `internal/analysis/service_test.go` — 新建，Service 测试
  * `internal/analysis/handler_test.go` — 新建，HTTP Handler 测试
* **关键签名** :
  ```go
  // internal/analysis/handler.go
  type Handler struct {
      svc    *Service
      logger *slog.Logger
  }

  func NewHandler(svc *Service, logger *slog.Logger) *Handler
  func (h *Handler) RegisterRoutes(mux *http.ServeMux)
  // POST /api/sessions/{id}/analyze  → handleAnalyze
  // GET  /api/sessions/{id}/report   → handleGetReport
  // GET  /api/sessions/{id}/reports  → handleListReports

  // internal/config/config.go 新增
  type AnalysisConfig struct {
      AutoTrigger bool `yaml:"auto_trigger"`
  }
  ```
* **验收标准** : `go build ./...` 通过；3 个路由正确注册；main.go 完整接线；AnalysisConfig 默认 AutoTrigger=true；POST /analyze 返回 201；GET /report 返回最新报告/404；GET /reports 返回历史列表；active 会话返回 409；全部单元测试通过
* **验证步骤** : `go build ./...` + `go test ./...`（V8/V10/V12）
* **测试要求** : L2，V1-V12 全部闭环
* **mapping_ids** : [V8, V10, V12]
* **依赖 / Wave** : wave-4，依赖 T3
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-task-3.json -> baseline/post-task-4.json`

#### 执行日志

| Task | 状态 | 涉及文件 | 验证证据 | 备注 |
|------|------|----------|----------|------|
| T1 | done | `internal/store/analysis.go`, `internal/store/schema.go`, `internal/store/db.go` | go build 成功 | AnalysisReport CRUD + Schema 迁移新增 2 列 |
| T2 | done | `internal/analysis/engine.go` | go build 成功 | 分析 Prompt 构造 + LLM 调用 + JSON 解析容错 + Markdown 渲染 |
| T3 | done | `internal/analysis/service.go`, `internal/analysis/errors.go`, `internal/session/service.go` | go build 成功 | 分析生命周期编排 + OnSessionEndFunc 回调 |
| T4 | done | `internal/analysis/handler.go`, `internal/config/config.go`, `cmd/server/main.go`, `internal/store/analysis_test.go`, `internal/analysis/engine_test.go`, `internal/analysis/service_test.go`, `internal/analysis/handler_test.go` | go build + 94 tests pass + go vet clean | 3个 HTTP 端点 + 配置 + 接线 + 全部测试 |
