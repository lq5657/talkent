---
change_id: llm-client
status: done
depends_on:
  - scaffold-project
parallel_safe: true
branch: feat/llm-client
created: 2026-04-25
updated: 2026-04-25
complexity: medium
min_validation_level: L2
---

### LLM 客户端抽象

#### 0.1 需求收敛记录

`cc-new-project 项目定义` → `MVP 路线图 Phase 0 P0` → `LLM 客户端抽象`（收敛方式：`backlog`）

#### 1. 背景与目标

Talkent Phase 0 需要 LLM 能力支撑对话生成和分析报告。本次 change 目标：在 `internal/llm/` 下实现 LLM 客户端模块，封装 OpenAI 兼容协议的 chat/completions 调用，为后续 `session` 和 `analysis` 模块提供统一的 LLM 交互接口。

#### 1.0 路线图对齐

- 属于 Phase 0 第二个 change，依赖 `scaffold-project`（已完成）
- 是 `chat-session` 和 `analysis-engine` 的前置依赖
- 与推荐 backlog 一致，未偏离

#### 1.1 本次不做

- 多 provider 同时配置和运行时切换
- 流式输出（streaming chat completion）
- 对话记忆管理（`internal/memory/`）
- 会话管理（`internal/session/`）
- 分析引擎（`internal/analysis/`）
- HTTP API handler（由后续 change 负责）

#### 2. 代码现状（Research Findings）

- `internal/llm/` 目录不存在（scaffold-project 未创建该包的源文件）
- `config.LLMConfig` 已定义 Provider/BaseURL/APIKey/Model/Timeout 字段
- `config.yaml` 已有 DeepSeek 的配置模板
- 环境变量 `TALKENT_LLM_API_KEY` 覆盖已实现
- `sashabaranov/go-openai` SDK 未引入 go.mod

#### 3. 功能点

* [ ] **Client 接口定义**：`Chat(ctx, messages, opts)` 方法，领域类型 `ChatMessage`/`ChatResponse`
* [ ] **go-openai 实现**：`openai.DefaultConfig` + `BaseURL` 适配非 OpenAI 端点
* [ ] **配置集成**：复用 `config.LLMConfig`，在 `cmd/server/main.go` 初始化 LLM 客户端
* [ ] **错误处理与重试**：429/500 自动重试（最多 3 次，指数退避），401 不重试，其他记录
* [ ] **日志与脱敏**：调用入口/结果/失败日志，不含 API Key
* [ ] **单元测试**：httptest 模拟 LLM API，覆盖正常调用、错误处理、重试、日志脱敏

#### 4. 业务规则

- 使用 `sashabaranov/go-openai` SDK 作为底层 HTTP 客户端
- 通过 `openai.DefaultConfig` + `config.BaseURL` 适配非 OpenAI provider（DeepSeek、通义千问等）
- 单活跃 provider：配置文件中只有一个 provider 生效，切换靠改配置重启
- 复用已有 `config.LLMConfig`，不新增配置结构
- 所有 LLM 调用必须透传 `context.Context`，支持超时和取消
- LLM API 调用失败需按错误类型分类处理：429 重试、500 重试、401 不重试、其他记录
- API Key 不得出现在日志中

#### 5. 数据变更

- **是否涉及 migration**：否
- 本次不涉及数据库 schema 变更

#### 6. 接口变更

- **是否涉及对外契约变更**：否
- **兼容性分类**：compatible_addition（新增内部模块，不修改现有对外接口）

| 操作 | 接口 | 方法 | 变更内容 | 兼容性 |
|------|------|------|----------|--------|
| 新增 | `llm.Client` | `Chat` | 内部接口，chat/completions 调用 | 新增 |
| 修改 | `cmd/server/main.go` | `main` | 新增 LLM 客户端初始化 | 兼容 |

#### 7. 影响范围

### 7.1 In-Scope

- `internal/llm/` 包：Client 接口定义、go-openai 实现、类型定义、单元测试
- `go.mod` / `go.sum`：新增 `sashabaranov/go-openai` 依赖
- `cmd/server/main.go`：集成 LLM 客户端初始化

### 7.2 Out-of-Scope

（同 §1.1）

#### 7.3 配置变更

- **是否涉及配置项或环境变量变更**：否（复用已有 `config.LLMConfig`）
- **配置来源**：YAML 配置文件 + 环境变量覆盖（已有，无需修改）
- **敏感配置**：`llm.api_key` 通过环境变量 `TALKENT_LLM_API_KEY` 注入，不硬编码

#### 8. 风险与关注点

- **外部依赖**：`sashabaranov/go-openai` 为社区维护库，需确认 API 稳定性
- **跨模型兼容性**：非 OpenAI provider 对 OpenAI 兼容协议的实现程度可能不一致

#### 8.1 日志与可观测性

- **是否新增运行时日志点**：是
- **涉及入口**：`OpenAIClient.Chat()` 调用
- **使用的 logger**：`log/slog`
- **关键字段**：provider、model、消息数、token 用量、错误类型、HTTP 状态码
- **日志落点**：随应用统一 logger
- **日志格式**：时间（微秒）、等级、消息、文件名:行号、函数/方法名

#### 9. 测试策略

- **测试范围**：`internal/llm/` 包
- **最低验证等级**：L2（Unit/Package）
- **验证证据要求**：`go test ./internal/llm/...` 全部通过
- **测试方式**：httptest 模拟 LLM API 端点

#### 9.1 需求-验证映射

| 编号 | 需求项 / 风险点 | 最低验证等级 | 证据类型 | 建议验证动作 | 对应 Task | 闭环状态 |
|------|------------------|--------------|----------|--------------|-----------|----------|
| V1 | Client 接口定义与 go-openai 实现可编译 | L2 | package | `go build ./...` | Task 1 | apply-covered |
| V2 | ChatCompletion 调用正确传递 messages/model/temperature | L2 | unit | httptest 模拟 LLM API 验证请求参数 | Task 4 | apply-covered |
| V3 | 自定义 BaseURL 生效（可连接非 OpenAI 端点） | L2 | unit | httptest 服务器验证请求 URL | Task 4 | apply-covered |
| V4 | 错误分类处理（429/500 重试，401 不重试，其他记录） | L2 | unit | httptest 返回不同状态码验证重试行为 | Task 4 | apply-covered |
| V5 | API Key 不出现在日志中 | L2 | unit | 检查日志输出不含 API Key | Task 4 | apply-covered |
| V6 | context.Context 透传与超时控制 | L2 | unit | context 取消时立即返回错误 | Task 4 | apply-covered |

#### 9.2 发布与回滚

低风险新增模块，直接代码提交 + git revert 即可回滚。

- **发布方式**：直接提交
- **回滚路径**：代码回滚（git revert），无数据库/配置迁移
- **发布后观察窗口**：合并后确认 `go build` 和 `go test` 通过

#### 10. 待澄清

- [x] 多模型策略：用户选择单活跃 provider
- [x] HTTP 客户端选型：用户选择 sashabaranov/go-openai

#### 11. 方案比较

| 维度 | 方案 A：sashabaranov/go-openai | 方案 B：标准库 net/http |
|------|-------------------------------|----------------------|
| 类型安全 | SDK 类型完整 | 需手写 request/response 类型 |
| 维护成本 | 依赖社区更新 | 零依赖，完全自主 |
| 兼容性 | 可能与部分 provider 不完全兼容 | 完全可控 |
| 选中 | ✅ | ❌ |

选择方案 A：用户明确选择，SDK 提供类型安全和错误分类，减少手写代码量。

#### 12. 技术决策

| 决策 | 选择 | 放弃的方案 | 原因 |
|------|------|-----------|------|
| 多模型策略 | 单活跃 provider | 多 provider 可选 | MVP 最简，用户确认 |
| HTTP 客户端 | sashabaranov/go-openai | 标准库 net/http | 用户选择；SDK 类型安全，支持自定义 BaseURL |
| 接口设计 | 领域类型 + Client 接口解耦 | 直接暴露 go-openai 类型 | 便于后续替换实现或 mock 测试 |
| 重试策略 | 指数退避，最多 3 次 | 固定间隔 / 无重试 | 429/500 场景下退避比固定间隔更合理 |

#### 13. 确认记录（HARD-GATE）

* **confirmed_at**：2026-04-25
* **confirmed_by**：lq5657
* **confirmed_spec_revision**：`spec.md` @ 2026-04-25，6 个功能点，V1-V6 验证映射（7 列），接口+go-openai 实现+错误重试+单元测试
* **confirmed_tasks_revision**：`tasks.md` @ 2026-04-25，4 个 task，4 个 wave
* **confirmed_scope**：`internal/llm/` Client 接口+go-openai 实现+错误处理/重试+单元测试；`cmd/server/main.go` LLM 客户端初始化；不含对话/分析/会话/记忆/前端
* **accepted_risks**：sashabaranov/go-openai 为社区维护库；非 OpenAI provider 兼容性可能不一致
* **human_review_required**：false
* **human_review_status**：not_required
