---
change_id: analysis-engine
status: done
depends_on:
  - scaffold-project
  - llm-client
  - role-and-goal
  - chat-session
parallel_safe: false
branch: feat/analysis-engine
created: 2026-04-27
updated: 2026-04-29
complexity: medium
min_validation_level: L2
---

### 分析引擎

#### 0.1 需求收敛记录

`cc-new-project 项目定义` → `MVP 路线图 Phase 0 P0` → `分析引擎：多维度分析报告生成，核心差异化`（收敛方式：`backlog`）

#### 1. 背景与目标

Talkent 的核心链路是"角色设定 → 对话训练 → 分析报告"。前四个 change 已完成角色设定（`role-and-goal`）、对话会话（`chat-session`）和记忆管理。本次 change 目标：在 `internal/analysis/` 下实现多维度分析引擎——从已结束的对话中提取各维度评分、评语和改进建议，生成 JSON 结构化数据 + Markdown 报告，为 Phase 1 Web 前端的报告展示提供数据基础。

这是 Phase 0 最后一个 change，完成后 Talkent 将跑通"设定→对话→分析"的完整链路。

#### 1.0 路线图对齐

- 属于 Phase 0 第五个（最后一个）change，依赖 `scaffold-project`（done）、`llm-client`（done）、`role-and-goal`（done）、`chat-session`（done）
- 是 `web-frontend` 的前置依赖
- 与推荐 backlog 一致，未偏离

#### 1.1 本次不做

- WebSocket / Streaming 输出
- 报告对比/趋势分析（跨会话对比）
- PDF / Word 导出
- 报告编辑/修订
- 自定义维度权重
- 分析结果缓存/预计算
- 前端报告页面展示

#### 2. 代码现状（Research Findings）

- `internal/analysis/` 目录存在但为空，无源码
- `internal/llm/` 已实现 `Client` 接口（`Chat(ctx, messages, opts)`），可直接调用
- `internal/role/` 已实现 `Dimension{Name, Description}` 模型和 `DimensionsForType()` 查表
- `internal/session/` 已实现 `Service`（创建、对话、结束会话）和 `GetSession()`
- `internal/store/session.go` 已实现 `SessionStore.GetSession()`、`ListMessages()`
- `internal/store/schema.go` 中 `analysis_reports` 表已有 `id`、`session_id`、`dimension_results`、`created_at` 字段
- 现有表缺少 `markdown_content` 和 `model_used` 列，需新增
- `internal/config/config.go` 无分析相关配置
- `cmd/server/main.go` 的依赖注入链：config→db→llm→roleSvc→sessionStore→memory→sessionSvc

#### 3. 功能点

* [ ] **Analysis Store CRUD**：`internal/store/analysis.go`，AnalysisReport 读写操作
* [ ] **Analysis Engine**：`internal/analysis/engine.go`，结构化分析 Prompt 构造、LLM 调用、JSON 解析容错、Markdown 渲染
* [ ] **Analysis Service**：`internal/analysis/service.go`，分析生命周期编排（触发、校验、持久化）
* [ ] **手动触发 API**：`POST /api/sessions/{id}/analyze`
* [ ] **自动触发钩子**：会话结束时自动触发分析（通过回调解耦）
* [ ] **报告查询 API**：`GET /api/sessions/{id}/report` 获取最新报告，`GET /api/sessions/{id}/reports` 列出历史
* [ ] **Schema 迁移**：新增 `markdown_content`、`model_used` 列到 `analysis_reports` 表
* [ ] **配置项**：`analysis.auto_trigger`（是否自动触发）
* [ ] **main.go 接线**：AnalysisStore → Engine → AnalysisService → AnalysisHandler → 路由注册 + 自动触发钩子

#### 4. 业务规则

- 仅 `completed` 状态的会话可触发分析；`active` 会话返回 409
- 分析维度来自会话创建时设定的 `dimensions` 字段
- LLM 分析使用一次调用输出所有维度结果（结构化 JSON）
- 分析 Prompt 格式：角色设定 + 场景 + 对话原文 + 分析维度列表 + JSON 输出格式要求
- JSON 解析失败时进行一次重试（重试 Prompt 强调格式要求）；仍失败则返回 500 并记录原始 LLM 输出
- 每次分析生成独立报告记录（允许重复分析），`analysis_reports` 表按 `created_at` 降序排列
- 报告包含 JSON 结构化数据（维度结果数组）+ Markdown 渲染全文
- 自动触发：会话结束时若 `analysis.auto_trigger = true`，自动触发分析；自动触发失败不影响会话结束
- `model_used` 记录生成报告时使用的 LLM 模型名
- 报告查询：`GET /report` 返回最新一份，`GET /reports` 返回全部历史

#### 4.1 API 兼容性分类

| 端点 | 类型 | 说明 |
|------|------|------|
| `POST /api/sessions/{id}/analyze` | compatible_addition | 新增端点 |
| `GET /api/sessions/{id}/report` | compatible_addition | 新增端点 |
| `GET /api/sessions/{id}/reports` | compatible_addition | 新增端点 |

#### 4.2 安全考量

- 用户输入（对话内容）经 LLM Prompt 传入：需与 System Prompt 严格分离
- LLM API Key 不在日志中输出：已有 `llm-client` 保证
- LLM 原始输出仅在解析失败时记录到日志，用于排障；需脱敏处理（截断超长内容）
- 无资金、权限相关高风险点

#### 5. 接口契约

##### POST /api/sessions/{id}/analyze — 触发分析

**Response (201):**
```json
{
  "report_id": 1,
  "session_id": "uuid",
  "dimensions": [
    {
      "name": "论点清晰度",
      "description": "核心论点是否明确、易懂",
      "score": 8,
      "comment": "论点表达较为清晰...",
      "suggestions": ["可以在开头更明确地点出核心论点"]
    }
  ],
  "markdown": "# 对话分析报告\n\n## 会话信息\n...",
  "model_used": "gpt-4o",
  "created_at": "2026-04-27T12:00:00Z"
}
```

- 对话仍在进行时返回 409
- 会话不存在返回 404
- LLM 调用/解析失败返回 500

##### GET /api/sessions/{id}/report — 获取最新报告

**Response (200):**
```json
{
  "report_id": 1,
  "session_id": "uuid",
  "dimensions": [...],
  "markdown": "...",
  "model_used": "gpt-4o",
  "created_at": "2026-04-27T12:00:00Z"
}
```

- 无报告时返回 404

##### GET /api/sessions/{id}/reports — 列出历史报告

**Response (200):**
```json
{
  "reports": [
    {
      "report_id": 2,
      "created_at": "2026-04-27T14:00:00Z",
      "model_used": "gpt-4o"
    },
    {
      "report_id": 1,
      "created_at": "2026-04-27T12:00:00Z",
      "model_used": "gpt-4o"
    }
  ]
}
```

##### 错误码

| HTTP | 场景 |
|------|------|
| 404 | 会话不存在 / 报告不存在 |
| 409 | 会话仍在进行中（非 completed） |
| 500 | LLM 调用失败 / JSON 解析重试后仍失败 |

#### 6. 数据模型

**Schema 迁移**：在 `analysis_reports` 表新增两列：

| 列 | 类型 | 约束 | 说明 |
|----|------|------|------|
| `markdown_content` | TEXT | NOT NULL DEFAULT '' | Markdown 渲染全文 |
| `model_used` | TEXT | NOT NULL DEFAULT '' | 使用的 LLM 模型名 |

迁移方式：通过 `store.Open()` 中的 DDL 扩展（`ALTER TABLE` 在 SQLite 中对新增列可用）。

现有列保持不变：
- `id`：INTEGER PRIMARY KEY AUTOINCREMENT
- `session_id`：TEXT NOT NULL，外键关联 sessions
- `dimension_results`：TEXT NOT NULL DEFAULT '[]'，JSON 格式存储维度结果
- `created_at`：DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP

#### 7. 配置项

| 配置名 | 默认值 | 必填 | 说明 | 环境变量 |
|--------|--------|------|------|----------|
| `analysis.auto_trigger` | true | 否 | 会话结束时是否自动触发分析 | `TALKENT_ANALYSIS_AUTO_TRIGGER` |

#### 8. 可观测性

- 分析触发：INFO（session_id, trigger: manual/auto）
- LLM 分析调用：INFO（session_id, model, dimension_count）
- JSON 解析失败第一次：WARN（session_id, retrying）
- JSON 解析重试后仍失败：ERROR（session_id, raw_output 截断至 500 字符）
- 报告生成完成：INFO（session_id, report_id, dimension_count, model_used）
- 自动触发失败不影响会话结束：WARN（session_id, error）

#### 9. 风险与缓解

| 风险 | 等级 | 缓解 |
|------|------|------|
| LLM 返回格式不稳定 | 中高 | 结构化 Prompt + JSON Schema 约束 + 一次重试 + 失败时记录原始输出 |
| 分析 Prompt token 超限 | 中 | 长对话截断策略：保留最近 N 轮完整对话 + 早期对话摘要 |
| 分析质量不够 | 中 | Prompt 精细设计 + 明确评分标准 + 具体示例 |
| 自动触发与手动触发并发 | 低 | SQLite 串行化；同一会话的并发分析各生成独立报告 |

#### 10. 影响范围

| 影响模块 | 变更类型 | 说明 |
|----------|----------|------|
| `internal/store/analysis.go` | 新增 | AnalysisReport CRUD |
| `internal/analysis/engine.go` | 新增 | 分析 Prompt 构造 + LLM 调用 + JSON 解析 + Markdown 渲染 |
| `internal/analysis/service.go` | 新增 | 分析生命周期编排 |
| `internal/analysis/handler.go` | 新增 | HTTP API |
| `internal/store/schema.go` | 修改 | 新增 2 列 DDL |
| `internal/session/service.go` | 修改 | 新增 OnSessionEndFunc 回调支持 |
| `internal/config/config.go` | 修改 | 新增 AnalysisConfig |
| `cmd/server/main.go` | 修改 | 依赖注入 + 自动触发钩子接线 |

#### 11. 拆分理由

本 change 有一个明确用户目标（"对已完成的对话生成多维度分析报告"），一个验收故事，一个验证集群。涉及 Store + Engine + Service + Handler 四个层面，但它们服务于同一个分析链路，拆开无法独立验收。复杂度 **medium**。

#### 12. 验证映射

| 编号 | 需求项 / 风险点 | 最低验证等级 | 证据类型 | 建议验证动作 | 对应 Task | 闭环状态 |
|------|------------------|--------------|----------|--------------|-----------|----------|
| V1 | Analysis Store CRUD | L2 | package | Report 读写单元测试 | T1 | apply-covered |
| V2 | 分析 Prompt 正确构造 | L2 | unit | 验证 Prompt 包含角色+场景+对话+维度+JSON 格式要求 | T2 | apply-covered |
| V3 | LLM 一次调用全维度分析 | L2 | unit | mock LLM 验证调用参数和返回解析 | T2 | apply-covered |
| V4 | JSON 解析失败重试 | L2 | unit | mock LLM 第一次返回非法 JSON，第二次返回合法 JSON | T2 | apply-covered |
| V5 | 重试仍失败返回错误 | L2 | unit | mock LLM 两次都返回非法 JSON，验证错误和日志 | T2 | apply-covered |
| V6 | Markdown 报告正确渲染 | L2 | unit | 验证 Markdown 包含标题+维度评分+评语+建议 | T2 | apply-covered |
| V7 | 仅 completed 会话可分析 | L2 | unit | active 会话返回 409 | T3 | apply-covered |
| V8 | 手动触发 API | L2 | unit | httptest 验证 analyze 端点 | T4 | apply-covered |
| V9 | 自动触发钩子 | L2 | unit | 会话结束时验证回调被调用 | T3 | apply-covered |
| V10 | 报告查询 API | L2 | unit | httptest 验证 report/reports 端点 | T4 | apply-covered |
| V11 | Schema 迁移正确执行 | L2 | unit | 验证新列存在且可读写 | T1 | apply-covered |
| V12 | main.go 接线后服务启动正常 | L1 | build | `go build ./...` 通过 | T4 | apply-covered |

#### 13. 确认记录（HARD-GATE）

* **confirmed_at**：2026-04-27
* **confirmed_by**：lq5657
* **confirmed_spec_revision**：`spec.md` @ 2026-04-27，9 个功能点，V1-V12 验证映射（7 列），分析触发+Engine+Service+Handler+Schema迁移+配置+单元测试
* **confirmed_tasks_revision**：`tasks.md` @ 2026-04-27，4 个 task，4 个 wave
* **confirmed_scope**：`internal/store/analysis.go` AnalysisReport CRUD；`internal/store/schema.go` 新增 markdown_content/model_used 列；`internal/analysis/engine.go` 分析 Prompt+LLM 调用+JSON 解析容错+Markdown 渲染；`internal/analysis/service.go` 分析生命周期编排；`internal/analysis/handler.go` 3个HTTP端点；`internal/session/service.go` OnSessionEndFunc 回调；`internal/config/config.go` AnalysisConfig；`cmd/server/main.go` 依赖注入+钩子接线；不含 WebSocket/流式/报告对比/PDF导出/前端展示
* **accepted_risks**：LLM JSON 格式不稳定（重试+截断日志）；长对话 token 超限（复用窗口截断）；分析质量不够（Prompt 精细设计）；自动与手动并发（SQLite 串行化）
* **human_review_required**：false
* **human_review_status**：not_required

#### 14. 回滚

- 新增代码：`internal/store/analysis.go`、`internal/analysis/engine.go`、`internal/analysis/service.go`、`internal/analysis/handler.go`
- 修改代码：`internal/store/schema.go`（DDL 扩展）、`internal/session/service.go`（回调支持）、`internal/config/config.go`、`cmd/server/main.go`
- Schema 迁移：新增列有默认值，回滚时删除新增列即可（SQLite 支持 `ALTER TABLE DROP COLUMN` 从 3.35.0）
- 回滚方式：删除新增文件，还原修改文件，删除新增列
