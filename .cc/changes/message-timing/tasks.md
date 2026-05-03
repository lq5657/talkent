---
change_id: message-timing
created: 2026-05-03
updated: 2026-05-03
---

### 任务拆分 — 对话页消息时间显示

#### 前置条件

* [x] `spec.md` 已确认且 `status = propose`
* [x] `depends_on` 为空，无前置变更依赖

#### 依赖 / Wave 总览

```
Wave 1: Task 1 (backend) → Task 2 (frontend, depends on Task 1)
```

#### 变更影响概览

##### 文件变更清单

| 文件 | 操作 | 涉及 Task | 说明 |
|------|------|-----------|------|
| `internal/session/service.go` | 修改 | Task 1 | ChatResult 新增时间戳字段 |
| `internal/session/handler.go` | 修改 | Task 1 | chatResponse 新增时间戳字段 |
| `web/src/api/client.ts` | 修改 | Task 2 | ChatResponse 接口新增时间戳字段 |
| `web/src/views/ChatView.vue` | 修改 | Task 2 | Message 接口 + sendMessage 逻辑更新 |
| `web/src/components/MessageBubble.vue` | 修改 | Task 2 | 新增时间信息展示 |

##### 受影响接口 / 调用方

| 接口 / 函数 / 入口 | 变更类型 | 上游调用方 | 下游依赖 | 涉及 Task |
|--------------------|----------|------------|----------|-----------|
| `POST /api/sessions/:id/chat` | compatible_addition | ChatView.vue sendMessage() | 无 | Task 1, Task 2 |
| `session.Service.Chat()` | 返回值扩展 | session.Handler.handleChat | llm.Client, memory.Manager | Task 1 |
| `chatResponse` | 字段新增 | handleChat → JSON encode | 前端 ChatResponse 类型 | Task 1, Task 2 |

##### 构建系统变更

无

#### Spec 覆盖映射

| Spec 章节 / 映射编号 | 覆盖 Task | 说明 |
|----------------------|-----------|------|
| V1: Chat API 返回时间戳 | Task 1 | 后端 L2 package test |
| V2: 消息时间信息展示 | Task 2 | 前端 L4 manual |

#### Task 1: 后端 ChatResult / chatResponse 新增时间戳字段

* **目标**: Chat API 响应中返回用户消息和 AI 回复的创建时间
* **不包含范围**: 不修改 store 层、不修改数据库 schema
* **涉及文件**: `internal/session/service.go`, `internal/session/handler.go`
* **上下游 Context**: ChatResult 由 Service.Chat() 返回 → Handler.handleChat() 构造 chatResponse → JSON 序列化返回前端
* **关键签名**:
  - `ChatResult` 新增 `UserMessageCreatedAt time.Time`, `AssistantMessageCreatedAt time.Time`
  - `chatResponse` 新增 `UserMessageCreatedAt string` (json:"user_message_created_at"), `AssistantMessageCreatedAt string` (json:"assistant_message_created_at")
* **验收标准**:
  - ChatResult 包含两个时间戳字段
  - chatResponse JSON 包含 `user_message_created_at` 和 `assistant_message_created_at`
  - `go test ./internal/session/` 通过
* **验证步骤**: `go test ./internal/session/ -v`
* **渐进可验证要求**: 修改后 `go build ./...` 通过 → 运行 handler 测试确认 JSON 序列化包含新字段
* **测试要求**: 更新或确认现有 handler 测试覆盖新字段
* **依赖 / Wave**: Wave 1，无前置依赖
* **回退方式**: revert service.go 和 handler.go 的改动
* **完成后状态**: `todo`
* **Baseline / Delta（按需）**:
* **对应 commit（按需）**:

#### Task 2: 前端类型、ChatView 和 MessageBubble 更新

* **目标**: 对话页每条消息显示开始时间、结束时间、持续时间
* **不包含范围**: 不修改消息存储/加载逻辑（刷新后消息不持久化是已知行为）
* **涉及文件**: `web/src/api/client.ts`, `web/src/views/ChatView.vue`, `web/src/components/MessageBubble.vue`
* **上下游 Context**: ChatView.sendMessage() 调用 api.chat() → 解析 ChatResponse → 推入 messages 数组 → MessageBubble 渲染
* **关键签名**:
  - `ChatResponse` 新增 `user_message_created_at: string`, `assistant_message_created_at: string`
  - `Message` 新增 `timestamp: Date`
  - `MessageBubble` props 新增 `startTime?: Date`, `endTime: Date`
* **验收标准**:
  - 每条用户消息下方显示 AI 回复到达时间（开始）、发送时间（结束）、耗时
  - 每条 AI 回复下方显示用户消息发送时间（开始）、AI 回复到达时间（结束）、耗时
  - 首条消息开始时间和持续时间显示 "—"
  - 持续时间 >= 60s: "X分Y秒"，< 60s: "Y秒"
* **验证步骤**: 浏览器访问对话页，发送消息，检查时间显示
* **渐进可验证要求**: 构建前端 → 启动服务 → 浏览器验证
* **测试要求**: 手工浏览器验证（L4 manual）
* **依赖 / Wave**: Wave 2，依赖 Task 1 完成
* **回退方式**: revert 三个前端文件的改动
* **Baseline / Delta（按需）**:
* **对应 commit（按需）**:
* **完成后状态**: `todo`
