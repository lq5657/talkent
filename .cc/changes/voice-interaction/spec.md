---
change_id: voice-interaction
status: done
depends_on: []
parallel_safe: true
branch: feat/voice-interaction
created: 2026-05-19
updated: 2026-05-20
complexity: medium
proposal_profile: standard
---

### 流式文字 + 浏览器语音交互

#### 0.1 需求收敛记录

`支持用户通过手机移动端和角色进行语音交互` → `STT 方案：浏览器 SpeechRecognition`、`TTS 方案：浏览器 SpeechSynthesis`、`交互模式：流式文字 + 录播式语音` → `语音输入填入输入框手动发送 + AI 回复自动朗读 + LLM 回复 SSE 流式显示`

收敛方式：两轮交互澄清（STT/TTS/交互模式/移动端范围 + 自动发送/自动朗读）

#### 0.2 Proposal Profile 与逐节确认

| Profile | 使用条件 | 文档要求 |
|---------|----------|----------|
| standard | 涉及外部契约变更（SSE endpoint）、多模块（llm/session/frontend）、流式架构 | 完整 spec + tasks + 验证映射 |

#### 1. 背景与目标

当前 Talkent 仅支持文字输入/输出。在手机移动端场景下，用户期望通过语音与 AI 角色对话，提升交互自然度和便捷性。

目标：在现有 Chat 页面基础上，增加浏览器原生语音输入（SpeechRecognition）和语音输出（SpeechSynthesis），同时将 LLM 回复改为 SSE 流式显示（打字效果），提升响应感知速度。

#### 1.0 路线图对齐

无。本 change 为独立功能增强。

#### 1.1 本次不做

- 服务端 STT/TTS（全部使用浏览器内置 API）
- 真正的实时流式语音对话（WebRTC/WebSocket/WebSocket）
- 全站移动端响应式改造、PWA
- Setup/Report 页面的语音功能
- 分析报告流式生成
- 多语言语音/国际化
- 语音活动检测 (VAD)
- 非流式 Chat endpoint 不变，保持向后兼容

#### 2. 代码现状（Research Findings）

##### 2.1 相关入口与链路

| 入口 | 路径 | 当前行为 |
|------|------|----------|
| Chat API | `POST /api/sessions/:id/chat` | 同步请求-响应，JSON body `{content}` → `{reply, round_info, timestamps}` |
| LLM Client 接口 | `internal/llm/client.go` | 单一方法 `Chat()`，同步返回完整文本 |
| OpenAI 实现 | `internal/llm/openai.go` | 调用 `CreateChatCompletion`（非流式），go-openai v1.41.2 已内置 `CreateChatCompletionStream` |
| 前端 ChatView | `web/src/views/ChatView.vue` | `sendMessage()` 一个 `await api.chat()` 获取完整回复 |
| 前端 API Client | `web/src/api/client.ts` | `fetch()` JSON 请求/响应，无流式方法 |
| 前端 ChatInput | `web/src/components/ChatInput.vue` | `<textarea>` + 发送按钮 |
| 前端 MessageBubble | `web/src/components/MessageBubble.vue` | 文字渲染 + 时间戳显示 |

##### 2.2 现有实现

- go-openai v1.41.2 已支持 `CreateChatCompletionStream`、`CreateSpeech`、`CreateTranscription`
- 浏览器 `SpeechRecognition` API 在 Chrome/Android 上可用，Safari/iOS 部分支持
- 浏览器 `SpeechSynthesis` API 在所有主流浏览器上可用
- 前端 `package.json` 无音频/media 依赖（仅有 vue/vue-router/marked/highlight.js/dompurify）
- 后端 `go.mod` 无需新增依赖

##### 2.3 发现与风险

| 发现 | 风险等级 | 处置 |
|------|----------|------|
| SpeechRecognition 仅在 Chrome/Android 上好用，Safari/iOS 支持有限 | 中 | Feature detection，仅在 API 可用时显示录音按钮 |
| SpeechSynthesis 声音因 OS/浏览器而异，不可控音色 | 低 | 接受为 MVP 级别，不强制统一音色 |
| SSE 连接在移动网络切换（Wi-Fi ↔ 蜂窝）时可能断开 | 中 | 前端增加断线重试 + 错误提示 |
| 当前 Chat 架构全链路非流式，需从 LLM Client → Service → Handler → 前端逐层打通 | 高 | 作为核心架构变更，分 5 个 task 逐层实现 |

#### 3. 功能点

- [ ] F1 语音输入：ChatInput 增加录音按钮，浏览器 SpeechRecognition API 识别语音为文字，结果填入输入框，用户手动点击发送
- [ ] F2 语音输出：AI 消息气泡增加播放按钮，浏览器 SpeechSynthesis API 将文字转为语音，回复完成后自动朗读，可随时重听
- [ ] F3 流式文字：LLM 回复通过 SSE 逐 token 推送至前端，前端实时渲染（打字效果）
- [ ] F4 语音状态 UI：录音中/识别中/播放中的视觉状态指示，移动端友好的按钮尺寸（≥44px 触摸目标）

#### 4. 业务规则

- 录音按钮仅在浏览器支持 `SpeechRecognition` API 时显示（`webkitSpeechRecognition` 或 `SpeechRecognition`）
- 播放按钮始终显示在 AI 消息上
- AI 回复流式完成后自动触发 TTS 朗读
- 用户在 TTS 播放中点击播放按钮 → 停止当前播放
- 用户在 TTS 未播放时点击播放按钮 → 开始朗读
- SSE 连接断开时，前端显示"连接中断"提示并允许重试
- 流式 Chat 和非流式 Chat 共享同一 session 状态校验逻辑（active/completed/round limit）

#### 5. 数据变更

- **是否涉及 migration**：否
- 无数据库 schema 变更

#### 6. 接口变更

- **是否涉及对外契约变更**：是
- **兼容性分类**：compatible_addition（新增 SSE endpoint，现有 POST /chat 不变）
- **客户端/消费者影响**：前端从非流式切换到流式（新 endpoint），旧 endpoint 保留
- **迁移路径**：前端切换调用新 SSE endpoint，旧 endpoint 保留向后兼容
- **回滚影响**：前端回退到 `POST /api/sessions/:id/chat` 非流式即可

| 操作 | 接口 | 方法 | 变更内容 | 兼容性 |
|------|------|------|----------|--------|
| 新增 | `/api/sessions/{id}/chat/stream` | GET | SSE 流式 Chat，query param `content`，返回 `text/event-stream` | compatible_addition |
| 调整 | `/api/sessions/{id}/chat` | POST | response 新增 `user_message_created_at`、`assistant_message_created_at` 字段（message-timing change 配套） | compatible_addition |

#### 7. 影响范围

| 模块 | 文件 | 改动类型 |
|------|------|----------|
| LLM Client | `internal/llm/client.go` | 接口扩展：新增 `ChatStream` 方法 + `StreamChunk` 类型 |
| LLM Client | `internal/llm/openai.go` | 实现 `ChatStream`：调用 `CreateChatCompletionStream` |
| Session Service | `internal/session/service.go` | 新增 `ChatStream` 方法：持久化 + 流式 LLM 调用 |
| Session Handler | `internal/session/handler.go` | 新增 SSE handler |
| Server Entry | `cmd/server/main.go` | 注册 SSE route |
| 前端 API | `web/src/api/client.ts` | 新增 `chatStream` 方法 |
| ChatView | `web/src/views/ChatView.vue` | 流式消息渲染 + 自动播放逻辑 |
| ChatInput | `web/src/components/ChatInput.vue` | 新增录音按钮 + SpeechRecognition |
| MessageBubble | `web/src/components/MessageBubble.vue` | 新增播放按钮 + SpeechSynthesis |

#### 7.1 配置变更

- **是否涉及配置项或环境变量变更**：否
- 所有语音功能使用浏览器内置 API，无需服务端配置

#### 8. 风险与关注点

| 类型 | 描述 | 处理方式 |
|------|------|----------|
| 兼容性 | SpeechRecognition 浏览器覆盖率有限 | Feature detection，不支持的浏览器隐藏录音按钮 |
| 稳定性 | SSE 连接在移动网络不稳定 | 前端断线重试 + 错误提示 |
| 体验 | SpeechSynthesis 音色因设备而异 | MVP 接受，后续可升级为服务端 TTS |
| 架构 | 流式链路从 LLM → 前端首次打通 | 分层实现，每层独立验证 |

#### 8.1 日志与可观测性

- **是否新增运行时日志点**：是（SSE 流式 Chat 的 session 生命周期日志）
- **使用的 logger**：slog（现有 TextHandler）
- **关键字段**：request_id, session_id
- **日志落点**：stdout 或配置的日志文件

#### 9. 测试策略

- **测试范围**：LLM Client 流式接口、Session Service 流式方法、SSE Handler、前端语音 UI
- **最低验证等级**：L2（package test）+ L4（manual browser）
- **验证证据要求**：go test 通过（L2）、手工移动端浏览器验证语音输入/输出（L4）

#### 9.1 需求-验证映射

| 编号 | 需求项 / 风险点 | 最低验证等级 | 证据类型 | 建议验证动作 | 对应 Task | 闭环状态 |
|------|------------------|--------------|----------|--------------|-----------|----------|
| V1 | LLM ChatStream 流式输出正确 | L2 | package | `go test ./internal/llm/` | Task 1 | todo |
| V2 | Session ChatStream 持久化 + 轮数控制 | L2 | package | `go test ./internal/session/` | Task 2 | todo |
| V3 | SSE endpoint 正确推送 token | L2 | package | `go test ./internal/session/` SSE handler | Task 3 | todo |
| V4 | 语音输入/输出 + 流式文字 UI | L4 | manual | 移动端浏览器手工验证全流程 | Task 4, Task 5 | todo |

#### 9.2 发布与回滚

- **发布方式**：直接发布（单服务，无灰度）
- **Feature Flag / Kill Switch**：无
- **回滚路径**：前端回退到 `POST /api/sessions/:id/chat` 非流式 + 移除语音按钮；后端 SSE endpoint 可保留不调用
- **发布后观察窗口**：验证 SSE 连接稳定性 + 语音功能可用性

#### 10. 待澄清

无。所有关键决策已通过交互澄清。

#### 10.1 风险决策（需用户选择）

| 决策风险 | 可选处理路径 | 推荐路径 | 用户选择 / 状态 |
|----------|--------------|----------|-----------------|
| SpeechRecognition 浏览器兼容性 | 仅 Chrome 支持 / 服务端 fallback / 混合 | Feature detection，不支持时隐藏 | Feature detection，不支持时隐藏 ✓ |
| SpeechSynthesis 音色不一致 | 接受 / 服务端 TTS | 接受为 MVP 级别 | 接受为 MVP 级别 ✓ |
| 自动发送 vs 手动发送 | 自动发送 / 填入输入框手动发送 | 自动发送 | 填入输入框手动发送 ✓ |
| 自动朗读 vs 手动播放 | 自动朗读 / 手动点击播放 | 自动朗读 | 自动朗读 ✓ |

#### 11. 方案比较

| 方案 | 是否采用 | 适用前提 | 采用 / 放弃原因 |
|------|----------|----------|-----------------|
| 浏览器内置语音 API | 采用 | 浏览器支持 SpeechRecognition + SpeechSynthesis | 零后端改动、零 API 成本、零新依赖 |
| 服务端 STT/TTS（go-openai） | 放弃 | 需要跨浏览器一致体验 | API 成本、延迟增加、不符合轻量 MVP 定位 |
| 真正实时语音对话（WebRTC） | 放弃 | 需要 ChatGPT 高级语音模式体验 | 架构改动巨大、复杂度高、与 MVP 目标不匹配 |
| 非流式文字（保持现状） | 放弃 | 无流式需求 | 用户选择了流式方案，感知响应更快 |

#### 12. 技术决策

| 决策 | 选择 | 放弃的方案 | 原因 |
|------|------|-----------|------|
| 流式方案 | SSE (Server-Sent Events) | WebSocket, 轮询 | 单向推送够用、HTTP 原生支持、前端 `EventSource`/`fetch` ReadableStream 简单 |
| STT API | 浏览器 SpeechRecognition | 服务端 Whisper | 零后端、免费、Chrome/Android 主要目标平台 |
| TTS API | 浏览器 SpeechSynthesis | 服务端 TTS | 零后端、免费、所有主流浏览器支持 |
| LLM 流式 | go-openai CreateChatCompletionStream | 自建 streaming | go-openai 已内置、无新依赖 |
| Chat endpoint | 新增独立 SSE endpoint | 改造现有 POST endpoint | 保持向后兼容、关注点分离 |

#### 13. 执行日志

| Task | 状态 | 实际改动文件 | Baseline / Delta | 备注 |
|------|------|-------------|------------------|------|
| Task 1: LLM ChatStream | done | client.go, openai.go, openai_test.go (+4 mock stubs) | commit 2e9cca9 | go test ./internal/llm/ 25 passed |
| Task 2: Session ChatStream | done | service.go, service_test.go | commit 512483c | go test ./internal/session/ 5 new streaming tests passed |
| Task 3: SSE Handler | done | handler.go, handler_test.go | commit 201b472 | go test ./internal/session/ 4 SSE handler tests passed |
| Task 4: Frontend Streaming | done | client.ts, ChatView.vue | commit e450e74 | vue-tsc + vite build passed, 115 Go tests |
| Task 5: Frontend Voice UI | done | ChatInput.vue, MessageBubble.vue | commit 07b9561 | vue-tsc + vite build passed |

#### 14. 审查结论

待实现完成后由 `cc-review` 填充。

#### 15. 确认记录（HARD-GATE）

- **confirmed_at**：2026-05-20
- **confirmed_by**：user
- **confirmed_spec_revision**：v1
- **confirmed_tasks_revision**：v1
- **confirmed_scope**：confirmed — 用户确认进入 cc-apply
- **resolved_risk_decisions**：confirmed — 用户确认进入 cc-apply
- **accepted_residual_risks**：confirmed — 用户确认进入 cc-apply
- **human_review_required**：false
- **human_review_status**：not_required
