### 变更日志 — 流式文字 + 浏览器语音交互

#### 时间线

| 时间 | 阶段 | 事件 | 备注 |
|------|------|------|------|
| 2026-05-19 | propose | cc-propose 创建 change，两轮交互澄清需求 | STT/TTS/交互模式/移动端范围 + 自动发送/朗读 |
| 2026-05-19 | propose | HARD-GATE：用户选择"暂停，等待澄清" | spec/tasks 已写入，auto-validation 全部通过，等待用户恢复 |
| 2026-05-20 | propose | 恢复 cc-propose，HARD-GATE 确认通过 | auto-validation 全部 PASSED，用户确认进入 cc-apply |
| 2026-05-20 | apply | Task 1-5 全部完成，promote 到 review | 5 commits, 115 Go tests, frontend build ok, cc-verify+delta passed |

#### 技术决策

| 决策 | 选择 | 放弃的方案 | 原因 |
|------|------|-----------|------|
| 流式方案 | SSE | WebSocket, 轮询 | 单向推送够用、HTTP 原生支持 |
| STT | 浏览器 SpeechRecognition | 服务端 Whisper | 零后端、免费 |
| TTS | 浏览器 SpeechSynthesis | 服务端 TTS | 零后端、免费 |
| LLM 流式 | go-openai CreateChatCompletionStream | 自建 | 已内置、无新依赖 |

#### Git / 验证记录

| 时间 | 分支/提交动作 | 影响范围 | 验证等级 | 备注 |
|------|---------------|----------|----------|------|
| 2026-05-20 | 2e9cca9 feat(llm) | llm client/interface/tests | L2 | 25 tests passed (4 new ChatStream) |
| 2026-05-20 | 512483c feat(session) | session service/tests | L2 | 5 new ChatStream tests passed |
| 2026-05-20 | 201b472 feat(session) | SSE handler/tests | L2 | 4 SSE handler tests passed |
| 2026-05-20 | e450e74 feat(frontend) | client.ts, ChatView.vue | L4 (build) | vue-tsc + vite build passed |
| 2026-05-20 | 07b9561 feat(frontend) | ChatInput.vue, MessageBubble.vue | L4 (build) | vue-tsc + vite build passed |

#### 踩坑记录

| 问题 | 原因 | 解决方案 | 已沉淀？ |
|------|------|----------|----------|

#### 知识候选 / 发现（按归档确认）

| 关键词 | 一句话结论 | 出处 | 建议落点 | 类型 | 复利判断 | 处理结果 |
|--------|------------|------|----------|------|----------|----------|
