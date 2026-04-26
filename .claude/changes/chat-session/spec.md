---
change_id: chat-session
status: review
depends_on:
  - scaffold-project
  - llm-client
  - role-and-goal
parallel_safe: false
branch: feat/chat-session
created: 2026-04-26
updated: 2026-04-26
complexity: medium
min_validation_level: L2
---

### 对话会话管理

#### 0.1 需求收敛记录

`cc-new-project 项目定义` → `MVP 路线图 Phase 0 P0` → `对话会话管理：创建、对话、记忆管理、结束`（收敛方式：`backlog`）

#### 1. 背景与目标

Talkent 的核心链路是"角色设定 → 对话训练 → 分析报告"。本次 change 目标：在 `internal/session/` 和 `internal/memory/` 下实现对话会话生命周期管理——创建会话（注入角色设定）、多轮对话（记忆管理 + LLM 调用）、结束会话（手动或自动），为后续分析引擎提供完整对话数据。

#### 1.0 路线图对齐

- 属于 Phase 0 第四个 change，依赖 `scaffold-project`（done）、`llm-client`（done）、`role-and-goal`（done）
- 是 `analysis-engine` 的前置依赖
- 与推荐 backlog 一致，未偏离

#### 1.1 本次不做

- 分析引擎与报告生成（`internal/analysis/`，属于 `analysis-engine` change）
- WebSocket / Streaming（MVP 使用同步 HTTP）
- 会话列表/搜索/删除（MVP 不需要）
- 消息编辑/重发
- 多用户/权限控制
- 前端页面
- 摘要持久化（MVP 每轮重新生成，保持简单）

#### 2. 代码现状（Research Findings）

- `internal/session/` 和 `internal/memory/` 目录存在但为空，无源码
- `internal/llm/` 已实现 `Client` 接口和 `OpenAIClient`，可直接调用
- `internal/role/` 已实现 `Service`（RecommendGoals, RecommendDimensions, DeriveDimensions）和 `Handler`
- `internal/store/schema.go` 中 `sessions` 表已有 `role_config`、`goals`、`dimensions`、`status`、`round_limit` 字段
- `internal/store/schema.go` 中 `messages` 表已有 `session_id`、`role`、`content`、`sequence_num` 字段
- `internal/store/` 仅有 `Open()` 和 `Schema`，无 CRUD 方法
- `internal/server/server.go` 支持 `registerRoutes` 回调注册路由
- `internal/config/` 无会话/记忆相关配置

#### 3. 功能点

* [ ] **Session Store CRUD**：`internal/store/session.go`，Session 和 Message 的读写操作
* [ ] **Memory Manager**：`internal/memory/manager.go`，滑动窗口 + LLM 摘要压缩 + 降级策略
* [ ] **Session Service**：`internal/session/service.go`，会话生命周期编排（创建、对话、结束）
* [ ] **System Prompt 构造**：从角色描述、场景、目标、维度构造系统提示词
* [ ] **对话流程**：用户消息 → 记忆上下文构造 → LLM 调用 → 消息持久化 → 轮次检查
* [ ] **会话结束**：手动结束 API + round_limit 自动结束
* [ ] **HTTP API**：创建会话、发送消息、结束会话、获取会话信息
* [ ] **配置项**：`memory_window_size`（滑动窗口大小）
* [ ] **main.go 接线**：SessionStore → MemoryManager → SessionService → SessionHandler → 路由注册

#### 4. 业务规则

- 创建会话时接收角色描述、场景、角色类型、训练目标、分析维度、轮次上限，持久化到 sessions 表
- round_limit = 0 表示不限轮次（仅在用户手动结束时关闭）
- System Prompt 格式：角色设定 + 场景背景 + 训练目标 + 期望分析维度，作为每轮 LLM 调用的第一条消息
- 每轮计为 1 round（1 user message + 1 assistant reply = 1 round）
- 滑动窗口保留最近 N 条消息（默认 10 条），窗口外消息触发 LLM 摘要
- 摘要作为 System Prompt 的补充注入，窗口内消息始终使用原文
- 摘要失败时降级为仅窗口模式，不阻断对话
- 摘要不持久化，每轮重新生成
- 手动结束：`POST /api/sessions/{id}/end`
- 自动结束：round_limit 到达后自动标记 `completed`
- 已结束会话（status=completed）不允许继续发送消息（返回 409）
- 会话状态仅 `active` 和 `completed`

#### 4.1 API 兼容性分类

| 端点 | 类型 | 说明 |
|------|------|------|
| `POST /api/sessions` | compatible_addition | 新增端点，不影响现有 API |
| `POST /api/sessions/{id}/chat` | compatible_addition | 新增端点，不影响现有 API |
| `POST /api/sessions/{id}/end` | compatible_addition | 新增端点，不影响现有 API |
| `GET /api/sessions/{id}` | compatible_addition | 新增端点，不影响现有 API |

#### 4.2 安全考量

- 用户输入（角色描述、消息内容）经用户输入传入 LLM Prompt：需与 System Prompt 严格分离，防止 Prompt 注入
- LLM API Key 不在日志中输出：已有 `llm-client` 保证
- 无资金、权限相关高风险点

#### 5. 接口契约

##### POST /api/sessions — 创建会话

**Request:**
```json
{
  "role_description": "面试者",
  "scenario": "技术面试",
  "role_type": "structured_expression",
  "goals": [{"name": "逻辑条理性", "description": "观点是否有清晰的逻辑链条"}],
  "dimensions": [{"name": "论点清晰度", "description": "核心论点是否明确"}],
  "round_limit": 10
}
```

**Response (201):**
```json
{
  "session_id": "uuid",
  "status": "active",
  "round_limit": 10,
  "created_at": "2026-04-26T12:00:00Z"
}
```

##### POST /api/sessions/{id}/chat — 发送消息

**Request:**
```json
{
  "content": "你好，我是来面试的"
}
```

**Response (200):**
```json
{
  "reply": "你好，请坐...",
  "round_info": {
    "current": 1,
    "limit": 10,
    "is_last": false
  },
  "memory_source": "window"
}
```

- `memory_source`: `"window"` 或 `"summary+window"`
- 轮次到达上限时 `is_last: true`，同时自动结束会话

##### POST /api/sessions/{id}/end — 手动结束

**Response (200):**
```json
{
  "session_id": "uuid",
  "status": "completed",
  "final_round": 5
}
```

##### GET /api/sessions/{id} — 获取会话信息

**Response (200):**
```json
{
  "session_id": "uuid",
  "status": "active",
  "role_description": "面试者",
  "round_limit": 10,
  "current_round": 3,
  "message_count": 6,
  "created_at": "2026-04-26T12:00:00Z"
}
```

##### 错误码

| HTTP | 场景 |
|------|------|
| 400 | 请求体无效、必填字段缺失 |
| 404 | 会话不存在 |
| 409 | 会话已结束，不允许继续对话 |
| 500 | LLM 调用失败、内部错误 |

#### 6. 数据模型

使用现有 `sessions` 和 `messages` 表，不新增表或列。

- `sessions.role_config`：JSON `{ "description", "scenario", "type" }`
- `sessions.goals`：训练目标 JSON
- `sessions.dimensions`：分析维度 JSON
- `sessions.status`：`active` / `completed`
- `sessions.round_limit`：轮次上限，0 表示不限
- `messages`：每轮 2 条（user + assistant）

#### 7. 配置项

| 配置名 | 默认值 | 必填 | 说明 | 环境变量 |
|--------|--------|------|------|----------|
| `session.memory_window_size` | 10 | 否 | 滑动窗口保留的最近消息数 | `TALKENT_SESSION_MEMORY_WINDOW_SIZE` |

#### 8. 可观测性

- 会话创建：INFO（session_id, role_type）
- 对话请求：INFO（session_id, current_round, memory_source）
- LLM 调用失败：ERROR（session_id, error）
- 记忆摘要触发：INFO（session_id, overflow_count）
- 会话结束：INFO（session_id, final_round, trigger: manual/auto）

#### 9. 风险与缓解

| 风险 | 等级 | 缓解 |
|------|------|------|
| LLM 摘要质量不稳定 | 中 | 摘要仅补充上下文，窗口内原文保留；失败降级为仅窗口 |
| 长对话 token 超限 | 中 | 窗口大小可配置；token 估算与截断 |
| 并发请求轮次计数不准 | 低 | SQLite 单写锁串行化；后续换库需加锁 |

#### 10. 影响范围

| 影响模块 | 变更类型 | 说明 |
|----------|----------|------|
| `internal/store/session.go` | 新增 | Session/Message CRUD |
| `internal/memory/manager.go` | 新增 | 滑动窗口 + LLM 摘要 |
| `internal/session/service.go` | 新增 | 会话生命周期编排 |
| `internal/session/handler.go` | 新增 | HTTP API |
| `internal/config/config.go` | 修改 | 新增 SessionConfig |
| `cmd/server/main.go` | 修改 | 依赖注入与路由注册 |

#### 11. 拆分理由

本 change 有一个明确用户目标（"创建对话会话并进行多轮训练对话"），一个验收故事，一个验证集群。涉及 Store + Memory + Service + Handler 四个层面，但它们服务于同一个对话生命周期链路，拆开无法独立验收。复杂度 **medium**。

#### 12. 验证映射

| 编号 | 需求项 / 风险点 | 最低验证等级 | 证据类型 | 建议验证动作 | 对应 Task | 闭环状态 |
|------|------------------|--------------|----------|--------------|-----------|----------|
| V1 | Session Store CRUD | L2 | package | Session/Message 读写单元测试 | T1 | apply-covered |
| V2 | System Prompt 正确构造 | L2 | unit | 验证 Prompt 包含角色+场景+目标+维度 | T3 | apply-covered |
| V3 | 滑动窗口保留最近 N 条消息 | L2 | unit | 窗口内消息正确拼接 | T2 | apply-covered |
| V4 | LLM 摘要在窗口溢出时触发 | L2 | unit | 窗口溢出时调用 LLM 摘要 | T2 | apply-covered |
| V5 | 摘要失败时降级为仅窗口 | L2 | unit | mock LLM 失败后降级 | T2 | apply-covered |
| V6 | round_limit 到达后自动结束 | L2 | unit | 模拟达到上限后 status 变为 completed | T3 | apply-covered |
| V7 | 手动结束 API 正常工作 | L2 | unit | httptest 验证 end 端点 | T4 | apply-covered |
| V8 | 已结束会话拒绝新消息 | L2 | unit | completed 状态返回 409 | T3 | apply-covered |
| V9 | Handler 参数校验 | L2 | unit | 必填缺失返回 400 | T4 | apply-covered |
| V10 | 会话信息查询 | L2 | unit | httptest 验证 GET 端点 | T4 | apply-covered |
| V11 | main.go 接线后服务启动正常 | L1 | build | `go build ./...` 通过 | T4 | apply-covered |

#### 13. 确认记录（HARD-GATE）

* **confirmed_at**：2026-04-26
* **confirmed_by**：lq5657
* **confirmed_spec_revision**：`spec.md` @ 2026-04-26，9 个功能点，V1-V11 验证映射（7 列），会话创建+对话流程+记忆管理+会话结束+HTTP API+配置+单元测试
* **confirmed_tasks_revision**：`tasks.md` @ 2026-04-26，5 个 task，5 个 wave
* **confirmed_scope**：`internal/store/session.go` Session/Message CRUD；`internal/memory/manager.go` 滑动窗口+LLM摘要+降级；`internal/session/service.go` 会话生命周期编排；`internal/session/handler.go` 4个HTTP端点；`internal/config/config.go` SessionConfig；`cmd/server/main.go` 依赖注入；不含分析引擎/WebSocket/流式/前端/摘要持久化
* **accepted_risks**：LLM 摘要质量不稳定（降级兜底）；长对话 token 超限（窗口可配置+截断）；并发轮次计数（SQLite 串行化）
* **human_review_required**：false
* **human_review_status**：not_required

#### 14. 回滚

- 新增代码：`internal/store/session.go`、`internal/memory/manager.go`、`internal/session/service.go`、`internal/session/handler.go`
- 修改代码：`internal/config/config.go`、`cmd/server/main.go`
- 不涉及数据库 schema 变更
- 回滚方式：删除新增文件，还原修改文件
