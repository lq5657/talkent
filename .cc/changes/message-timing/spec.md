---
change_id: message-timing
status: done
depends_on: []
parallel_safe: true
branch: feat/message-timing
created: 2026-05-03
updated: 2026-05-03
complexity: medium
proposal_profile: micro
---

### 对话页消息时间显示

#### 1. 背景与目标

每次用户在对话页发送消息后，显示每条消息（用户消息和 AI 回复）的开始时间、结束时间和持续时间。帮助用户了解对话节奏和自己的思考速度。

#### 1.1 本次不做

- 页面刷新后恢复消息时间（当前无消息持久化加载）
- 消息时间在分析报告中的展示
- 时间字段的国际化

#### 2. 代码现状（Research Findings）

##### 2.1 相关入口与链路

- Chat API: `POST /api/sessions/:id/chat` → `session.Handler.handleChat` → `session.Service.Chat`
- 前端: `ChatView.vue` → `MessageBubble.vue`

##### 2.2 现有实现

- 后端 `store.Message` 已有 `CreatedAt time.Time`，但 Chat API 响应 (`chatResponse`) 未返回
- 前端 `Message` 类型仅 `{ role, content }`，无时间字段
- `MessageBubble` 仅显示角色标签 + 内容

##### 2.3 发现与风险

- Chat API 一次调用创建两条消息（user + assistant），需返回两个时间戳
- 首条消息无上一条消息，需特殊处理

#### 3. 功能点

* [ ] R1: Chat API 响应新增 `user_message_created_at` 和 `assistant_message_created_at` 字段
* [ ] R2: 前端每条消息气泡显示开始时间、结束时间、持续时间

#### 4. 业务规则

- 开始时间 = 上一条消息的创建时间
- 结束时间 = 本条消息的创建时间
- 持续时间 = 结束时间 − 开始时间
- 首条消息：仅显示结束时间，开始时间和持续时间显示 "—"
- 持续时间 >= 60s: 格式化为 `X分Y秒`；< 60s: `Y秒`
- 结束时间格式: `HH:mm:ss`

#### 5. 数据变更

* **是否涉及 migration**：否
* **migration / 脚本路径**：—
* **变更类型**：无
* **兼容窗口**：—
* **回滚路径**：—
* **数据回填方案**：—
* **幂等性与失败恢复**：—

| 操作 | 表名 | 字段/索引 | 说明 | 风险 |

#### 6. 接口变更

* **是否涉及对外契约变更**：是
* **兼容性分类**：compatible_addition
* **客户端/消费者影响**：前端需更新以使用新字段
* **迁移路径**：前端同步部署
* **回滚影响**：前端回退到旧版本忽略新字段

| 操作 | 接口 | 方法 | 变更内容 | 兼容性 |
|------|------|------|----------|--------|
| 新增字段 | /api/sessions/:id/chat | POST | 响应新增 user_message_created_at, assistant_message_created_at | compatible_addition |

#### 7. 影响范围

- `internal/session/service.go`: `ChatResult` 结构体
- `internal/session/handler.go`: `chatResponse` 结构体 + `handleChat`
- `web/src/api/client.ts`: `ChatResponse` 接口
- `web/src/views/ChatView.vue`: `Message` 接口 + `sendMessage()`
- `web/src/components/MessageBubble.vue`: 时间显示 UI

#### 7.1 配置变更

* **是否涉及配置项或环境变量变更**：否

#### 8. 风险与关注点

| 类型 | 描述 | 处理方式 |
|------|------|----------|
| 首条消息无上一条时间 | 第一条消息无法计算开始时间和持续时间 | 显示 "—" |
| 响应体积 | 新增 2 个 ISO 8601 字符串 | ~50 bytes，可忽略 |

#### 8.1 日志与可观测性

* **是否新增运行时日志点**：否

#### 9. 测试策略

* **测试范围**：后端 handler 测试验证 chatResponse 包含时间戳；前端手工验证时间显示
* **最低验证等级**：L2
* **验证证据要求**：`go test ./internal/session/` 通过；手工浏览器截图
* **若无法达到目标等级的替代方案**：—

#### 9.1 需求-验证映射

| 编号 | 需求项 / 风险点 | 最低验证等级 | 证据类型 | 建议验证动作 | 对应 Task | 闭环状态 |
|------|------------------|--------------|----------|--------------|-----------|----------|
| V1 | R1: Chat API 返回时间戳 | L2 | package | go test ./internal/session/ | Task 1 | apply-covered |
| V2 | R2: 消息时间信息展示 | L4 | manual | 浏览器验证对话页消息时间显示 | Task 2 | apply-covered |

#### 9.2 发布与回滚

* **发布方式**：直接发布
* **Feature Flag / Kill Switch**：无
* **回滚路径**：代码回滚
* **若无法直接回滚的原因**：—
* **发布后观察窗口**：—
* **失败触发条件**：—

#### 10. 待澄清

* [ ] 无

#### 10.1 风险决策（需用户选择）

| 决策风险 | 可选处理路径 | 推荐路径 | 用户选择 / 状态 |
|----------|--------------|----------|-----------------|
| 首条消息开始时间处理 | 显示"—" / 使用 session.created_at / 隐藏整行 | 显示"—" | resolved: 已实现为"—" |

#### 12. 技术决策

| 决策 | 选择 | 放弃的方案 | 原因 |
|------|------|-----------|------|
| 时间来源 | 后端 DB created_at | 前端 Date.now() | 服务端时间一致，避免客户端时钟偏差 |
| 开始时间定义 | 上一条消息的 created_at | 输入框 focus / 首次键入 | 用户确认 |

#### 13. 执行日志

| Task | 状态 | 实际改动文件 | Baseline / Delta | 备注 |
|------|------|-------------|------------------|------|
| Task 1 | done | service.go, handler.go | pre-apply.json → post-apply.json: unchanged-pass | L2 verified (go test) |
| Task 2 | done | client.ts, ChatView.vue, MessageBubble.vue | — | L4 pending (manual browser) |

#### 14. 审查结论

* **Stage 1 / Spec Compliance**：
* **Stage 2 / Code Quality**：
* **总体结论**：

#### 15. 确认记录（HARD-GATE）

* **confirmed_at**：2026-05-03
* **confirmed_by**：user
* **confirmed_spec_revision**：2026-05-03
* **confirmed_tasks_revision**：2026-05-03
* **confirmed_scope**：已确认：5 文件，2 tasks，L2+L4 验证
* **resolved_risk_decisions**：首条消息时间显示"—"，时间来源用后端 DB created_at
* **accepted_residual_risks**：无
* **human_review_required**：true
* **human_review_status**：approved
