---
change_id: role-and-goal
created: 2026-04-26
updated: 2026-04-26
---

### 任务拆分 — 角色设定与目标推荐

#### 前置条件

* [ ] `spec.md` 已确认且 `status = apply`
* [ ] HARD-GATE 已通过
* [ ] `scaffold-project` 已完成
* [ ] `llm-client` 已完成

#### Task 1: 角色数据模型与预置模板

* **目标** : 定义 `Role`、`TrainingGoal`、`Dimension` 领域类型，内置 `StructuredExpression` 模板与维度映射表
* **不包含范围** : 不实现 Service 逻辑（属于 T2/T3），不编写 Handler（属于 T4），不编写测试（属于 T5）
* **涉及文件** :
  * `internal/role/model.go` — 新建，领域类型定义
  * `internal/role/template.go` — 新建，预置模板与匹配函数
* **关键签名** :
  ```go
  type RoleType string
  const RoleTypeStructuredExpression RoleType = "structured_expression"

  type Role struct {
      Description string   `json:"description"`
      Scenario    string   `json:"scenario"`
      Type        RoleType `json:"type"`
  }

  type TrainingGoal struct {
      Name        string `json:"name"`
      Description string `json:"description"`
  }

  type Dimension struct {
      Name        string `json:"name"`
      Description string `json:"description"`
  }

  type RoleTemplate struct {
      Type           RoleType
      Keywords       []string
      DefaultGoals   []TrainingGoal
      DefaultDims    []Dimension
  }

  func MatchTemplate(desc string) (*RoleTemplate, bool)
  func DimensionsForType(rt RoleType) ([]Dimension, bool)
  ```
* **验收标准** : `go build ./...` 通过；`Role`/`TrainingGoal`/`Dimension` 类型已定义；`MatchTemplate` 按关键词匹配返回 `StructuredExpression` 模板；`DimensionsForType` 查表返回 5 个 MVP 维度
* **验证步骤** : `go build ./...`（V1, V2, V4）
* **测试要求** : L2，V1/V2/V4 在本 task 关闭为 `apply-covered`
* **mapping_ids** : [V1, V2, V4]
* **依赖 / Wave** : wave-1，无前置依赖
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/pre-apply.json -> baseline/post-task-1.json`

#### Task 2: 目标推荐 Service

* **目标** : 实现目标推荐逻辑（模板匹配 → LLM 生成兜底 → 用户确认）
* **不包含范围** : 不实现维度映射（属于 T3），不编写 Handler（属于 T4），不编写测试（属于 T5）
* **涉及文件** :
  * `internal/role/service.go` — 新建，Service 结构体与 RecommendGoals 方法
* **关键签名** :
  ```go
  type Service struct {
      llmClient llm.Client
      logger    *slog.Logger
  }

  func NewService(llmClient llm.Client, logger *slog.Logger) *Service
  func (s *Service) RecommendGoals(ctx context.Context, roleDesc string) ([]TrainingGoal, error)
  ```
* **验收标准** : `RecommendGoals` 先调 `MatchTemplate`，匹配到返回模板目标，匹配不到调 `llm.Client.Chat` 生成 JSON 格式目标列表；LLM Prompt 约束输出为 `{"goals": [...]}` 结构
* **验证步骤** : `go build ./...`（V3 编译级，完整功能验证在 T5）
* **测试要求** : L2，V3/V6 依赖 T5 单元测试闭环
* **mapping_ids** : [V3, V6]
* **依赖 / Wave** : wave-2，依赖 T1
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-task-1.json -> baseline/post-task-2.json`

#### Task 3: 维度映射 Service

* **目标** : 实现维度确定逻辑（查表优先 → LLM 推导兜底 → 用户最终决定）
* **不包含范围** : 不编写 Handler（属于 T4），不编写测试（属于 T5）
* **涉及文件** :
  * `internal/role/service.go` — 修改，追加 RecommendDimensions 和 DeriveDimensions 方法
* **关键签名** :
  ```go
  func (s *Service) RecommendDimensions(ctx context.Context, roleType RoleType, goals []TrainingGoal) ([]Dimension, error)
  func (s *Service) DeriveDimensions(ctx context.Context, roleDesc string, goals []TrainingGoal) ([]Dimension, error)
  ```
* **验收标准** : `RecommendDimensions` 先调 `DimensionsForType` 查表；`DeriveDimensions` 调 `llm.Client.Chat` 推导维度，Prompt 约束输出为 `{"dimensions": [...]}` 结构
* **验证步骤** : `go build ./...`（V5 编译级，完整功能验证在 T5）
* **测试要求** : L2，V5/V6 依赖 T5 单元测试闭环
* **mapping_ids** : [V5, V6]
* **依赖 / Wave** : wave-3，依赖 T1、T2
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-task-2.json -> baseline/post-task-3.json`

#### Task 4: HTTP API Handler 与路由注册

* **目标** : 实现角色设定、目标推荐、维度确定的 HTTP 端点，注册路由并接入 main.go
* **不包含范围** : 不编写测试（属于 T5）
* **涉及文件** :
  * `internal/role/handler.go` — 新建，HTTP Handler
  * `internal/server/server.go` — 修改，接收 role.Service 依赖，注册路由
  * `cmd/server/main.go` — 修改，初始化 role.Service 并传递给 server
* **关键签名** :
  ```go
  // internal/role/handler.go
  type Handler struct {
      svc    *Service
      logger *slog.Logger
  }

  func NewHandler(svc *Service, logger *slog.Logger) *Handler
  func (h *Handler) RegisterRoutes(mux *http.ServeMux)

  // 端点：
  // POST /api/roles/recommend-goals    — 接收角色描述，返回推荐目标
  // POST /api/roles/recommend-dimensions — 接收角色类型+目标，返回推荐维度；mode=derive 触发 LLM 推导
  ```
* **验收标准** : `server.New()` 接收 `*role.Handler` 参数；`main.go` 初始化 `role.NewService(llmClient, logger)` 并创建 Handler；两个 API 端点可访问；Prompt 注入防护：用户输入仅出现在 User Prompt
* **验证步骤** : `go build ./...`（V7/V8/V9）
* **测试要求** : L2，V7/V8 依赖 T5 单元测试闭环；V9 在本 task 关闭为 `apply-covered`
* **mapping_ids** : [V7, V8, V9]
* **依赖 / Wave** : wave-4，依赖 T2、T3
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-task-3.json -> baseline/post-task-4.json`

#### Task 5: 单元测试

* **目标** : 为 T1-T4 所有功能点编写单元测试
* **不包含范围** : 不测试真实 LLM API 调用（属于 L4 集成验证），不测试其他模块
* **涉及文件** :
  * `internal/role/model_test.go` — 新建，模型与模板测试
  * `internal/role/service_test.go` — 新建，Service 逻辑测试
  * `internal/role/handler_test.go` — 新建，HTTP Handler 测试
* **关键签名** :
  ```go
  // model_test.go
  func TestMatchTemplate_Hit(t *testing.T)
  func TestMatchTemplate_Miss(t *testing.T)
  func TestDimensionsForType(t *testing.T)

  // service_test.go
  func TestRecommendGoals_TemplateMatch(t *testing.T)
  func TestRecommendGoals_LLMFallback(t *testing.T)
  func TestRecommendDimensions_TableLookup(t *testing.T)
  func TestDeriveDimensions_LLM(t *testing.T)

  // handler_test.go
  func TestRecommendGoalsHandler(t *testing.T)
  func TestRecommendDimensionsHandler(t *testing.T)
  func TestRecommendDimensionsHandler_DeriveMode(t *testing.T)
  func TestPromptInjection_UserInputNotInSystemPrompt(t *testing.T)
  ```
* **验收标准** : `go test ./internal/role/...` 全部通过；覆盖 V1-V8 所有验证项
* **验证步骤** : `go test ./internal/role/... -v`（V1-V8）
* **测试要求** : L2，V1-V8 在本 task 关闭为 `apply-covered`
* **mapping_ids** : [V1, V2, V3, V4, V5, V6, V7, V8]
* **依赖 / Wave** : wave-5，依赖 T1-T4
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-task-4.json -> baseline/post-task-5.json`

#### 执行日志

| Task | 状态 | 涉及文件 | 验证证据 | 备注 |
|------|------|----------|----------|------|
| T1 | done | internal/role/model.go, internal/role/template.go | go build PASSED | 模型+模板+匹配+维度查表 |
| T2 | done | internal/role/service.go | go build PASSED, go vet PASSED | 目标推荐：模板匹配+LLM兜底 |
| T3 | done | internal/role/service.go | go build PASSED, go vet PASSED | 维度映射：查表+LLM推导 |
| T4 | done | internal/role/handler.go, internal/server/server.go, cmd/server/main.go | go build PASSED, go vet PASSED | API端点+路由+接线 |
| T5 | done | internal/role/model_test.go, service_test.go, handler_test.go | go test 26 passed in role package | V1-V9全覆盖 |
