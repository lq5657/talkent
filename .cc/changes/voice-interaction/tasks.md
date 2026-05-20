---
change_id: voice-interaction
created: 2026-05-19
updated: 2026-05-19
---

### 任务拆分 — 流式文字 + 浏览器语音交互

#### 前置条件

- [x] `spec.md` 已确认且 `status = propose`
- [x] `depends_on` 为空，无前置变更依赖

#### 依赖 / Wave 总览

```
Wave 1 (后端流式): Task 1 → Task 2 → Task 3
Wave 2 (前端):     Task 4 (依赖 Task 3) ∥ Task 5 (独立，可与 Wave 1 并行)
```

#### 变更影响概览

##### 文件变更清单

| 文件 | 操作 | 涉及 Task | 说明 |
|------|------|-----------|------|
| `internal/llm/client.go` | 修改 | Task 1 | 新增 `ChatStream` 方法 + `StreamChunk` 类型 |
| `internal/llm/openai.go` | 修改 | Task 1 | 实现 `ChatStream` |
| `internal/session/service.go` | 修改 | Task 2 | 新增 `ChatStream` 方法 |
| `internal/session/handler.go` | 修改 | Task 3 | 新增 SSE handler |
| `cmd/server/main.go` | 修改 | Task 3 | 注册 SSE route |
| `web/src/api/client.ts` | 修改 | Task 4 | 新增 `chatStream` 方法 |
| `web/src/views/ChatView.vue` | 修改 | Task 4 | 流式消息渲染 + 自动播放 |
| `web/src/components/ChatInput.vue` | 修改 | Task 5 | 录音按钮 + SpeechRecognition |
| `web/src/components/MessageBubble.vue` | 修改 | Task 5 | 播放按钮 + SpeechSynthesis |

##### 受影响接口 / 调用方

| 接口 / 函数 / 入口 | 变更类型 | 上游调用方 | 下游依赖 | 涉及 Task |
|--------------------|----------|------------|----------|-----------|
| `llm.Client` interface | 扩展（新增方法） | session.Service | go-openai | Task 1 |
| `session.Service.Chat` | 新增同级方法 `ChatStream` | session.Handler | llm.Client, store | Task 2 |
| `POST /api/sessions/{id}/chat` | 保留不变 | 前端旧逻辑 | session.Service.Chat | — |
| `GET /api/sessions/{id}/chat/stream` | 新增 | 前端新逻辑 | session.Service.ChatStream | Task 3 |
| `api.chat()` | 保留不变 | ChatView (旧) | POST /chat | — |
| `api.chatStream()` | 新增 | ChatView (新) | GET /chat/stream | Task 4 |
| ChatInput `send` event | 保留不变 | ChatView.sendMessage | — | — |
| ChatInput 新增 mic button | 新增 | 用户操作 | SpeechRecognition API | Task 5 |
| MessageBubble 新增 play button | 新增 | 用户操作 | SpeechSynthesis API | Task 5 |

##### 构建系统变更

无。Go 和 npm 依赖均不变。

#### Spec 覆盖映射

| Spec 章节 / 映射编号 | 覆盖 Task | 说明 |
|----------------------|-----------|------|
| F3 流式文字 | Task 1, 2, 3, 4 | 后端 LLM → Service → SSE → 前端渲染 |
| F1 语音输入 | Task 5 | ChatInput 录音按钮 + SpeechRecognition |
| F2 语音输出 | Task 5 | MessageBubble 播放按钮 + SpeechSynthesis |
| F4 语音状态 UI | Task 5 | 录音/识别/播放状态指示 |
| V1 LLM ChatStream | Task 1 | package test |
| V2 Session ChatStream | Task 2 | package test |
| V3 SSE endpoint | Task 3 | package test |
| V4 语音 UI | Task 4, 5 | manual browser |

#### Task 1: LLM Client — ChatStream 接口与实现

- **目标**: 在 LLM Client 接口中新增流式 Chat 方法，并用 go-openai 实现
- **不包含范围**: Service 层流式逻辑、SSE handler、前端
- **涉及文件**: `internal/llm/client.go`, `internal/llm/openai.go`
- **上下游 Context**: 下游 go-openai `CreateChatCompletionStream`；上游 session.Service 将调用此方法
- **关键签名**:
  ```go
  // client.go — 新增类型
  type StreamChunk struct {
      Content string
      Done    bool
      Error   error
  }

  // client.go — Client 接口扩展
  type Client interface {
      Chat(ctx context.Context, messages []ChatMessage, opts *ChatOptions) (*ChatResponse, error)
      ChatStream(ctx context.Context, messages []ChatMessage, opts *ChatOptions) (<-chan StreamChunk, error)
  }
  ```
- **验收标准**:
  - `Client` 接口包含 `ChatStream` 方法
  - `openaiClient.ChatStream()` 调用 `CreateChatCompletionStream`，返回 channel
  - Channel 正确推送 `StreamChunk`（Content + Done=true on EOF）
  - 错误通过 `StreamChunk.Error` 传播
  - 现有 `Chat` 方法行为不变
- **验证步骤**: `go test ./internal/llm/`
- **渐进可验证要求**: 运行 `go test ./internal/llm/` 确认现有测试通过 + 新增流式测试
- **测试要求**: 新增 `TestChatStream` 测试用例（mock 或真实 LLM endpoint）
- **依赖 / Wave**: Wave 1，无前置
- **回退方式**: 删除 `ChatStream` 方法，恢复原始 interface（向后兼容，仅扩展）
- **完成后状态**: `done`
- **Baseline / Delta（按需）**:

#### Task 2: Session Service — ChatStream 方法

- **目标**: 在 Session Service 中新增 `ChatStream`，复用持久化逻辑 + 流式 LLM 调用
- **不包含范围**: HTTP handler、SSE 协议、路由注册
- **涉及文件**: `internal/session/service.go`
- **上下游 Context**: 上游 handler；下游 llm.Client.ChatStream、store、memory.Manager
- **关键签名**:
  ```go
  // service.go — ChatStream 返回 channel 而非单个 result
  func (s *Service) ChatStream(ctx context.Context, sessionID, userContent string) (<-chan llm.StreamChunk, error)
  ```
- **验收标准**:
  - 复用 `Chat()` 的校验逻辑（session 存在、active 状态、非空 content）
  - 复用 `Chat()` 的持久化逻辑（user message 写入 → 构建 context → 调用 LLM → assistant message 写入）
  - 流式推送 LLM token 到 channel
  - 流结束后 assistant message 已持久化（完整文本），再发送 Done chunk
  - 轮数上限触发时：先持久化 assistant message → UpdateSessionStatus(completed) → notifySessionEnd → 再发 Done
  - 错误时关闭 channel 并传播 error chunk
- **验证步骤**: `go test ./internal/session/`
- **渐进可验证要求**: 运行 `go test ./internal/session/` 确认现有测试 + 新流式测试通过
- **测试要求**: 新增 `TestChatStream` 测试用例（mock LLM client 返回模拟 stream）
- **依赖 / Wave**: Wave 1，依赖 Task 1
- **回退方式**: 删除 `ChatStream` 方法即可（不影响现有 `Chat`）
- **完成后状态**: `done`
- **Baseline / Delta（按需）**:

#### Task 3: SSE Handler + Route

- **目标**: 新增 SSE endpoint，将 ChatStream 的 token 通过 SSE 推送到前端
- **不包含范围**: 前端流式消费、语音 UI
- **涉及文件**: `internal/session/handler.go`, `cmd/server/main.go`
- **上下游 Context**: 上游 HTTP 请求；下游 session.Service.ChatStream
- **关键签名**:
  ```go
  // handler.go
  func (h *Handler) handleChatStream(w http.ResponseWriter, r *http.Request) {
      sessionID := r.PathValue("id")
      content := r.URL.Query().Get("content")
      // ... SSE 推送
  }
  ```
- **验收标准**:
  - Route `GET /api/sessions/{id}/chat/stream?content=...` 注册成功
  - SSE headers 正确设置：`Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`
  - 每个 `StreamChunk.Content` 作为 `data: {"token":"..."}` 推送
  - 流结束时发送 `data: {"done":true, "round_info":{...}, "timestamps":{...}}`
  - 错误时发送 `data: {"error":"..."}` 并关闭连接
  - 复用现有 Chat 的校验错误映射（404/409/400）
  - CORS headers 对 SSE endpoint 生效
- **验证步骤**: `go test ./internal/session/` + `curl` 手工验证 SSE 输出
- **渐进可验证要求**: 启动服务 → `curl -N "http://localhost:8080/api/sessions/{id}/chat/stream?content=hello"` 观察逐 token 推送
- **测试要求**: 新增 handler 测试（mock service 返回模拟 stream channel）
- **依赖 / Wave**: Wave 1，依赖 Task 2
- **回退方式**: 删除 route 注册 + handler 方法（不影响现有 POST /chat）
- **完成后状态**: `done`
- **Baseline / Delta（按需）**:

#### Task 4: Frontend Streaming

- **目标**: 前端 API client 增加流式请求 + ChatView 改为流式消息渲染 + 自动 TTS 触发
- **不包含范围**: 录音按钮、SpeechRecognition、播放按钮 UI（属于 Task 5）
- **涉及文件**: `web/src/api/client.ts`, `web/src/views/ChatView.vue`
- **上下游 Context**: 下游 `GET /api/sessions/{id}/chat/stream`；上游 ChatView.sendMessage()
- **关键签名**:
  ```typescript
  // client.ts
  async function* chatStream(sessionId: string, content: string): AsyncGenerator<ChatStreamEvent>
  
  // ChatStreamEvent = { type: 'token', token: string } 
  //                | { type: 'done', round_info: ..., timestamps: ... }
  //                | { type: 'error', error: string }
  ```
- **验收标准**:
  - `chatStream()` 使用 `fetch()` + `ReadableStream` 解析 SSE 事件
  - ChatView 在发送消息后切换到流式模式：AI 消息逐 token 追加显示
  - 流完成后消息气泡显示完整内容 + 时间戳
  - 流完成后自动触发 TTS 朗读（调用 `speechSynthesis.speak()`）
  - SSE 连接断开时显示"连接中断"提示 + 重试按钮
  - Round limit 触发时正确处理 auto-end + 跳转 report
  - 保留 `text` mode 回退能力（可通过变量切换回非流式 POST /chat）
  - 非流式 Chat 作为 SSE 失败的 fallback
- **验证步骤**: 启动 dev server → 移动端浏览器或 Chrome DevTools 模拟 → 发送消息 → 观察流式打字效果 → 确认 TTS 自动播放
- **渐进可验证要求**: SSE 连接建立 → token 推送 → 消息渲染 → TTS 触发
- **测试要求**: 手工验证（L4）
- **依赖 / Wave**: Wave 2，依赖 Task 3
- **回退方式**: 切换回 `api.chat()` 非流式调用
- **完成后状态**: `done`
- **Baseline / Delta（按需）**:

#### Task 5: Frontend Voice UI

- **目标**: ChatInput 增加录音按钮 + MessageBubble 增加播放按钮 + 语音状态管理
- **不包含范围**: 流式文字渲染（Task 4）、后端语音处理
- **涉及文件**: `web/src/components/ChatInput.vue`, `web/src/components/MessageBubble.vue`
- **上下游 Context**: SpeechRecognition API（录音→文字），SpeechSynthesis API（文字→语音）
- **关键签名**:
  ```typescript
  // ChatInput.vue — 新增
  const isRecording = ref(false)
  const isRecognizing = ref(false)
  const recognitionSupported = ref(false) // 'SpeechRecognition' in window || 'webkitSpeechRecognition' in window
  function startRecording() { /* new SpeechRecognition() → start() */ }
  function stopRecording() { /* recognition.stop() */ }
  // recognition.onresult → emit('update:modelValue', transcript)
  
  // MessageBubble.vue — 新增 props
  // autoPlay?: boolean
  const isPlaying = ref(false)
  function togglePlay() { /* speechSynthesis.speak() or .cancel() */ }
  ```
- **验收标准**:
  - ChatInput：当 `recognitionSupported=true` 时显示麦克风按钮
  - 点击麦克风 → 开始录音（按钮变红/脉冲动画）→ 停止 → 显示"识别中..." → 文字填入输入框
  - 识别错误时显示 toast 提示
  - MessageBubble：AI 消息上显示播放按钮
  - 点击播放 → TTS 朗读（按钮变暂停图标）→ 朗读结束（恢复播放图标）
  - 同时只有一个 TTS 在播放（新播放自动停止旧播放）
  - 移动端按钮触摸目标 ≥44px（Tailwind `min-w-[44px] min-h-[44px]`）
  - 录音按钮和播放按钮在不同状态下有清晰的视觉区分
- **验证步骤**: 移动端浏览器 → 创建会话 → 点击录音 → 说话 → 发送 → 观察 AI 回复自动朗读 → 点击播放按钮重听
- **渐进可验证要求**: 录音按钮显示 → 录音 → 识别成功 → 播放按钮显示 → 播放成功
- **测试要求**: 手工验证（L4），需在 Chrome/Android 上测试
- **依赖 / Wave**: Wave 2，独立（可与 Task 1-4 并行开发）
- **回退方式**: 移除语音按钮相关代码，恢复纯文字 ChatInput/MessageBubble
- **完成后状态**: `done`
- **Baseline / Delta（按需）**:
