---
change_id: role-and-goal
status: review
depends_on:
  - scaffold-project
  - llm-client
parallel_safe: false
branch: feat/role-and-goal
created: 2026-04-26
updated: 2026-04-26
complexity: medium
min_validation_level: L2
---

### 角色设定与目标推荐

#### 0.1 需求收敛记录

`cc-new-project 项目定义` → `MVP 路线图 Phase 0 P0` → `角色设定 → 训练目标推荐 → 分析维度确定`（收敛方式：`backlog`）

#### 1. 背景与目标

Talkent 的核心链路起点是"用户设定角色 → 系统推荐目标 → 确定分析维度"。本次 change 目标：在 `internal/role/` 下实现角色与目标模块，包括角色数据模型、预置角色模板、目标推荐（模板匹配 + LLM 生成兜底）、维度映射（查表优先 + LLM 推导兜底），以及对应的 HTTP API。

#### 1.0 路线图对齐

- 属于 Phase 0 第三个 change，依赖 `scaffold-project`（done）和 `llm-client`（done）
- 是 `chat-session` 和 `analysis-engine` 的前置依赖
- 与推荐 backlog 一致，未偏离

#### 1.1 本次不做

- 会话管理与对话流程（`internal/session/`）
- 对话记忆管理（`internal/memory/`）
- 分析引擎与报告生成（`internal/analysis/`）
- Web 前端页面
- 多角色原型体系扩展（MVP 只做"结构化表达/逻辑思维"一种原型）
- 用户体系与认证

#### 2. 代码现状（Research Findings）

- `internal/role/` 目录存在但为空，无源码
- `internal/llm/` 已实现 `Client` 接口和 `OpenAIClient`，可直接调用
- `internal/store/schema.go` 中 `sessions` 表已有 `role_config TEXT`、`goals TEXT`、`dimensions TEXT` 字段
- `internal/server/server.go` 当前仅有 `GET /health`，需要新增 API 路由
- `internal/config/` 无角色相关配置
- `cmd/server/main.go` 中 `llmClient` 已创建但 `_ = llmClient` 未接入路由

#### 3. 功能点

* [ ] **角色数据模型**：定义 `Role` 结构体（描述、场景、角色类型），持久化到 `sessions.role_config`
* [ ] **预置角色模板**：内置"结构化表达/逻辑思维"原型的模板，含默认目标与维度映射
* [ ] **目标推荐 — 模板匹配**：根据角色描述关键词匹配预置模板，返回对应训练目标列表
* [ ] **目标推荐 — LLM 生成**：模板匹配失败时，调用 LLM 根据角色描述生成训练目标
* [ ] **维度映射 — 查表优先**：角色类型→维度映射表，MVP 硬编码 5 个维度（论点清晰度、论证结构、口头禅检测、回应度、改进建议）
* [ ] **维度映射 — LLM 推导兜底**：用户对查表结果不满意时，调用 LLM 根据角色×目标推导维度列表
* [ ] **用户确认/补充**：目标和维度都支持用户追加、删除、修改后再确认
* [ ] **HTTP API**：角色设定 + 目标推荐 + 维度确定 + 用户确认 的 API 端点
* [ ] **main.go 接线**：将 `llmClient` 注入 role Service，注册 API 路由

#### 4. 业务规则

- 角色类型用枚举常量表示，MVP 阶段仅 `StructuredExpression`（结构化表达/逻辑思维）
- 预置模板包含角色类型、典型描述关键词、默认训练目标列表、默认分析维度列表
- 目标推荐流程：先按角色描述关键词匹配模板 → 匹配到则返回模板目标 → 匹配不到则 LLM 生成 → 用户确认/补充
- 维度确定流程：先查角色类型→维度映射表 → 用户不满意则 LLM 推导 → 用户最终决定
- 用户可以对推荐的目标和维度进行追加、删除、修改
- LLM 生成目标和推导维度时，使用结构化 Prompt 约束输出格式为 JSON
- 角色、目标、维度的最终确认结果持久化到 `sessions` 表对应字段

#### 4.1 角色类型与预置模板

| 角色类型 | 典型场景关键词 | 默认训练目标 | 默认分析维度 |
|----------|---------------|-------------|-------------|
| `StructuredExpression` | 面试、汇报、演讲、辩论、说服、表达 | 逻辑条理性、论证充分性、重点突出、语言精练 | 论点清晰度、论证结构、口头禅检测、回应度、改进建议 |

#### 4.2 API 兼容性分类

| 端点 | 类型 | 说明 |
|------|------|------|
| `POST /api/roles/recommend-goals` | compatible_addition | 新增端点，不影响现有 API |
| `POST /api/roles/recommend-dimensions` | compatible_addition | 新增端点，不影响现有 API |

#### 4.3 安全考量

- 角色描述经用户输入传入 LLM Prompt：需与 System Prompt 严格分离，防止 Prompt 注入
- LLM API Key 不在日志中输出：已有 `llm-client` 保证
- 无资金、权限、状态流转相关高风险点

#### 5. 影响范围

| 影响模块 | 变更类型 | 说明 |
|----------|----------|------|
| `internal/role/` | 新增 | 角色模型、模板、Service、Handler |
| `internal/server/server.go` | 修改 | 注册新 API 路由，接收 role Service 依赖 |
| `cmd/server/main.go` | 修改 | 将 `llmClient` 注入 role Service，传递给 server |

#### 6. 依赖

- `scaffold-project`（done）：项目结构、配置、SQLite、HTTP Server
- `llm-client`（done）：`Client` 接口，用于目标推荐和维度推导的 LLM 调用

#### 7. 拆分理由

本 change 有一个明确的用户目标（"设定角色后获得推荐目标和分析维度"），一个验收故事，一个验证集群。虽然涉及数据模型 + LLM 调用 + API 三个层面，但它们服务于同一个推导链路，拆开反而无法独立验收。复杂度评估为 **medium**。

#### 8. 验证映射

| 编号 | 需求项 / 风险点 | 最低验证等级 | 证据类型 | 建议验证动作 | 对应 Task | 闭环状态 |
|------|------------------|--------------|----------|--------------|-----------|----------|
| V1 | 角色数据模型定义与序列化 | L2 | unit | `Role`/`TrainingGoal`/`Dimension` 结构体 JSON 序列化 | T1 | todo |
| V2 | 预置模板匹配返回正确目标 | L2 | unit | `MatchTemplate` 输入含关键词的描述返回模板 | T1 | todo |
| V3 | LLM 生成目标（模板未匹配时） | L2 | unit | httptest mock LLM API 验证 JSON 目标列表 | T5 | todo |
| V4 | 维度查表映射正确 | L2 | unit | `DimensionsForType` 返回 5 个 MVP 维度 | T1 | todo |
| V5 | LLM 推导维度（用户不满意查表结果时） | L2 | unit | httptest mock LLM API 验证 JSON 维度列表 | T5 | todo |
| V6 | 用户确认/补充目标和维度 | L2 | unit | Service 接受用户追加/删除/修改 | T5 | todo |
| V7 | HTTP API 端点路由与请求/响应 | L2 | unit | httptest 验证 POST 请求路由与 JSON 响应 | T5 | todo |
| V8 | Prompt 注入防护（System/User 分离） | L2 | unit | 验证用户输入不出现在 System Prompt | T5 | todo |
| V9 | main.go 接线后服务启动正常 | L1 | build | `go build ./...` 通过 | T4 | todo |

#### 9. 确认记录（HARD-GATE）

* **confirmed_at**：2026-04-26
* **confirmed_by**：lq5657
* **confirmed_spec_revision**：`spec.md` @ 2026-04-26，9 个功能点，V1-V9 验证映射（7 列），角色模型+预置模板+目标推荐+维度映射+HTTP API+单元测试
* **confirmed_tasks_revision**：`tasks.md` @ 2026-04-26，5 个 task，5 个 wave
* **confirmed_scope**：`internal/role/` 角色模型+预置模板+目标推荐（模板匹配+LLM 兜底）+维度映射（查表+LLM 推导兜底）+HTTP Handler；`internal/server/server.go` 路由注册；`cmd/server/main.go` 依赖注入；不含会话/记忆/分析/前端
* **accepted_risks**：LLM 生成目标/推导维度的 Prompt 质量需后续调优；预置模板关键词覆盖度有限
* **human_review_required**：false
* **human_review_status**：not_required

#### 10. 回滚

- 本次只新增代码和修改 main.go/server.go 的依赖注入
- 不涉及数据库 schema 变更（sessions 表已有 role_config/goals/dimensions 字段）
- 回滚方式：删除 `internal/role/`，还原 main.go 和 server.go 的变更
