---
change_id: chat-session
created: 2026-04-26
updated: 2026-04-26
---

### 任务拆分 — 对话会话管理

#### 前置条件

* [ ] `spec.md` 已确认且 `status = apply`
* [ ] HARD-GATE 已通过
* [ ] `scaffold-project` 已完成
* [ ] `llm-client` 已完成
* [ ] `role-and-goal` 已完成

#### Task 1: Session Store CRUD

* **目标** : 实现 Session 和 Message 的数据库读写操作
* **不包含范围** : 不实现 Memory Manager（属于 T2），不实现 Service 逻辑（属于 T3），不编写测试（属于 T5）
* **涉及文件** :
  * `internal/store/session.go` — 新建，Session/Message 领域模型 + CRUD
* **关键签名** :
  ```go
  type Session struct {
      ID          string
      RoleConfig  string // JSON
      Goals       string // JSON
      Dimensions  string // JSON
      Status      string // "active" | "completed"
      RoundLimit  int
      CreatedAt   time.Time
      UpdatedAt   time.Time
  }

  type Message struct {
      ID          int64
      SessionID   string
      Role        string // "user" | "assistant"
      Content     string
      SequenceNum int
      CreatedAt   time.Time
  }

  type SessionStore struct {
      db *sql.DB
  }

  func NewSessionStore(db *sql.DB) *SessionStore
  func (s *SessionStore) CreateSession(ctx context.Context, session *Session) error
  func (s *SessionStore) GetSession(ctx context.Context, id string) (*Session, error)
  func (s *SessionStore) UpdateSessionStatus(ctx context.Context, id string, status string) error
  func (s *SessionStore) CreateMessage(ctx context.Context, msg *Message) error
  func (s *SessionStore) ListMessages(ctx context.Context, sessionID string) ([]Message, error)
  func (s *SessionStore) CountMessages(ctx context.Context, sessionID string) (int, error)
  ```
* **验收标准** : `go build ./...` 通过；Session CRUD 方法可正常读写 sessions 表；Message CRUD 方法可正常读写 messages 表
* **验证步骤** : `go build ./...`（V1 编译级，完整功能验证在 T5）
* **测试要求** : L2，V1 在 T5 单元测试闭环
* **mapping_ids** : [V1]
* **依赖 / Wave** : wave-1，无前置依赖
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/pre-apply.json -> baseline/post-task-1.json`

#### Task 2: Memory Manager — 滑动窗口 + LLM 摘要

* **目标** : 实现对话记忆管理：滑动窗口保留最近 N 条消息，窗口溢出时调用 LLM 生成摘要，摘要失败时降级为仅窗口
* **不包含范围** : 不实现 Session Service（属于 T3），不编写测试（属于 T5）
* **涉及文件** :
  * `internal/memory/manager.go` — 新建，Manager 结构体 + BuildContext + Summarize
* **关键签名** :
  ```go
  type Manager struct {
      windowSize int
      llmClient  llm.Client
      logger     *slog.Logger
  }

  func NewManager(windowSize int, llmClient llm.Client, logger *slog.Logger) *Manager
  func (m *Manager) BuildContext(ctx context.Context, systemPrompt string, messages []store.Message) ([]llm.ChatMessage, string, error)
  // 返回: LLM 消息列表, memory_source("window"|"summary+window"), error
  func (m *Manager) Summarize(ctx context.Context, messages []store.Message) (string, error)
  ```
* **验收标准** : `go build ./...` 通过；BuildContext 在消息数 <= windowSize 时直接拼接；消息数 > windowSize 时对窗口外消息调 LLM 摘要并注入 system prompt；摘要失败时降级为仅窗口 + memory_source="window"
* **验证步骤** : `go build ./...`（V3/V4/V5 编译级，完整功能验证在 T5）
* **测试要求** : L2，V3/V4/V5 在 T5 单元测试闭环
* **mapping_ids** : [V3, V4, V5]
* **依赖 / Wave** : wave-2，依赖 T1（Message 类型定义）
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-task-1.json -> baseline/post-task-2.json`

#### Task 3: Session Service — 会话生命周期编排

* **目标** : 实现会话创建、对话流程、会话结束的业务编排
* **不包含范围** : 不编写 Handler（属于 T4），不编写测试（属于 T5）
* **涉及文件** :
  * `internal/session/service.go` — 新建，Service 结构体 + CreateSession/Chat/EndSession/GetSession/BuildSystemPrompt
* **关键签名** :
  ```go
  type Service struct {
      store    *store.SessionStore
      memory   *memory.Manager
      llmClient llm.Client
      logger   *slog.Logger
  }

  type CreateSessionRequest struct {
      RoleDescription string          `json:"role_description"`
      Scenario        string          `json:"scenario"`
      RoleType        role.RoleType   `json:"role_type"`
      Goals           []role.TrainingGoal `json:"goals"`
      Dimensions      []role.Dimension    `json:"dimensions"`
      RoundLimit      int             `json:"round_limit"`
  }

  type ChatResult struct {
      Reply        string `json:"reply"`
      CurrentRound int    `json:"current_round"`
      RoundLimit   int    `json:"limit"`
      IsLast       bool   `json:"is_last"`
      MemorySource string `json:"memory_source"`
  }

  func NewService(store *store.SessionStore, memory *memory.Manager, llmClient llm.Client, logger *slog.Logger) *Service
  func (s *Service) CreateSession(ctx context.Context, req CreateSessionRequest) (*store.Session, error)
  func (s *Service) Chat(ctx context.Context, sessionID string, userContent string) (*ChatResult, error)
  func (s *Service) EndSession(ctx context.Context, sessionID string) (*store.Session, error)
  func (s *Service) GetSession(ctx context.Context, sessionID string) (*store.Session, error)
  func BuildSystemPrompt(desc, scenario string, goals []role.TrainingGoal, dims []role.Dimension) string
  ```
* **验收标准** : `go build ./...` 通过；CreateSession 生成 UUID 并持久化；BuildSystemPrompt 包含角色+场景+目标+维度；Chat 流程端到端正确；round_limit 到达自动结束；已结束会话 Chat 返回错误
* **验证步骤** : `go build ./...`（V2/V6/V8 编译级，完整功能验证在 T5）
* **测试要求** : L2，V2/V6/V8 在 T5 单元测试闭环
* **mapping_ids** : [V2, V6, V8]
* **依赖 / Wave** : wave-3，依赖 T1、T2
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-task-2.json -> baseline/post-task-3.json`

#### Task 4: HTTP Handler + 路由注册 + 启动集成 + 配置

* **目标** : 实现 4 个 HTTP 端点，注册路由，修改 main.go 依赖注入，新增 SessionConfig
* **不包含范围** : 不编写测试（属于 T5）
* **涉及文件** :
  * `internal/session/handler.go` — 新建，HTTP Handler + 4 个端点
  * `internal/config/config.go` — 修改，新增 SessionConfig + 环境变量覆盖
  * `cmd/server/main.go` — 修改，构造 SessionStore → MemoryManager → SessionService → SessionHandler → 路由注册
* **关键签名** :
  ```go
  // internal/session/handler.go
  type Handler struct {
      svc    *Service
      logger *slog.Logger
  }

  func NewHandler(svc *Service, logger *slog.Logger) *Handler
  func (h *Handler) RegisterRoutes(mux *http.ServeMux)
  // POST /api/sessions        → handleCreate
  // POST /api/sessions/{id}/chat → handleChat
  // POST /api/sessions/{id}/end  → handleEnd
  // GET  /api/sessions/{id}      → handleGet

  // internal/config/config.go
  type SessionConfig struct {
      MemoryWindowSize int `yaml:"memory_window_size"`
  }
  ```
* **验收标准** : `go build ./...` 通过；4 个路由正确注册；main.go 完整接线；SessionConfig 默认值 windowSize=10
* **验证步骤** : `go build ./...`（V7/V9/V10/V11）
* **测试要求** : L2，V7/V9/V10 在 T5 单元测试闭环；V11 在本 task 关闭为 `apply-covered`
* **mapping_ids** : [V7, V9, V10, V11]
* **依赖 / Wave** : wave-4，依赖 T3
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-task-3.json -> baseline/post-task-4.json`

#### Task 5: 单元测试

* **目标** : 为 T1-T4 所有功能点编写单元测试
* **不包含范围** : 不测试真实 LLM API 调用（属于 L4 集成验证），不测试其他模块
* **涉及文件** :
  * `internal/store/session_test.go` — 新建，Session/Message CRUD 测试
  * `internal/memory/manager_test.go` — 新建，Memory Manager 测试
  * `internal/session/service_test.go` — 新建，Service 逻辑测试
  * `internal/session/handler_test.go` — 新建，HTTP Handler 测试
* **关键签名** :
  ```go
  // store/session_test.go
  func TestCreateAndGetSession(t *testing.T)
  func TestUpdateSessionStatus(t *testing.T)
  func TestCreateAndListMessages(t *testing.T)

  // memory/manager_test.go
  func TestBuildContext_WithinWindow(t *testing.T)
  func TestBuildContext_OverflowTriggerSummary(t *testing.T)
  func TestBuildContext_SummaryFailureDegradation(t *testing.T)

  // session/service_test.go
  func TestCreateSession(t *testing.T)
  func TestBuildSystemPrompt(t *testing.T)
  func TestChat_EndToEnd(t *testing.T)
  func TestChat_RoundLimitAutoEnd(t *testing.T)
  func TestChat_CompletedSessionRejected(t *testing.T)
  func TestEndSession(t *testing.T)

  // session/handler_test.go
  func TestHandleCreate(t *testing.T)
  func TestHandleChat(t *testing.T)
  func TestHandleEnd(t *testing.T)
  func TestHandleGet(t *testing.T)
  func TestHandleCreate_MissingFields(t *testing.T)
  ```
* **验收标准** : `go test ./internal/store/... ./internal/memory/... ./internal/session/...` 全部通过；覆盖 V1-V10 所有验证项
* **验证步骤** : `go test ./internal/store/... ./internal/memory/... ./internal/session/... -v`（V1-V10）
* **测试要求** : L2，V1-V10 在本 task 关闭为 `apply-covered`
* **mapping_ids** : [V1, V2, V3, V4, V5, V6, V7, V8, V9, V10]
* **依赖 / Wave** : wave-5，依赖 T1-T4
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-task-4.json -> baseline/post-task-5.json`

#### 执行日志

| Task | 状态 | 涉及文件 | 验证证据 | 备注 |
|------|------|----------|----------|------|
| T1 | done | internal/store/session.go | go build PASSED | Session/Message CRUD |
| T2 | done | internal/memory/manager.go | go build PASSED, go vet PASSED | 滑动窗口+LLM摘要+降级 |
| T3 | done | internal/session/service.go, internal/session/errors.go | go build PASSED, go vet PASSED | 会话生命周期编排 |
| T4 | done | internal/session/handler.go, internal/config/config.go, cmd/server/main.go | go build PASSED, go vet PASSED | 4个HTTP端点+配置+接线 |
| T5 | done | internal/store/session_test.go, internal/memory/manager_test.go, internal/session/service_test.go, internal/session/handler_test.go | go test 22 passed (store 4, memory 4, session 14) | V1-V11全覆盖 |
