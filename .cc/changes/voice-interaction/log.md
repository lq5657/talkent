### 变更日志 — 流式文字 + 浏览器语音交互

#### 时间线

| 时间 | 阶段 | 事件 | 备注 |
|------|------|------|------|
| 2026-05-19 | propose | cc-propose 创建 change，两轮交互澄清需求 | STT/TTS/交互模式/移动端范围 + 自动发送/朗读 |
| 2026-05-19 | propose | HARD-GATE：用户选择"暂停，等待澄清" | spec/tasks 已写入，auto-validation 全部通过，等待用户恢复 |
| 2026-05-20 | propose | 恢复 cc-propose，HARD-GATE 确认通过 | auto-validation 全部 PASSED，用户确认进入 cc-apply |
| 2026-05-20 | apply | Task 1-5 全部完成，promote 到 review | 5 commits, 115 Go tests, frontend build ok, cc-verify+delta passed |
| 2026-05-20 | review | cc-review 完成：Stage1 PASS, Stage2 PASS, 4 findings | F1(%s JSON), F2(duplicate fallback), F3(goroutine ctx), F4(spec §6) |
| 2026-05-20 | fix | F1-F4 全部修复 | commits b399702, f00a963, dc91a24 |
| 2026-05-20 | archive | L4 验证发现 F5 (Flusher middleware), 修复后归档 | 3 knowledge entries 沉淀; commit e9c33c4, 50b9fd2 |

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
| 2026-05-20 | dbd486c review | review.md | — | Stage1 PASS, Stage2 PASS, 4 findings |
| 2026-05-20 | b399702 fix(F1) | handler.go | L2 | %s → %q JSON escaping |
| 2026-05-20 | f00a963 fix(F2) | ChatView.vue | L4 (build) | always pop before fallback |
| 2026-05-20 | dc91a24 fix(F3+F4) | service.go, spec.md | L2 | ctx.Err() check + spec §6 update |
| 2026-05-20 | aec5cc4 review | review.md, task-board | — | mark all findings fixed |
| 2026-05-20 | e9c33c4 fix(F5) | middleware.go + L4 screenshot | L2+L4 | Flusher fix + browser evidence |
| 2026-05-20 | 50b9fd2 archive | spec.md, review, knowledge/ | — | status=done, 3 knowledge entries |

#### 踩坑记录

| 问题 | 原因 | 解决方案 | 已沉淀？ |
|------|------|----------|----------|

#### 知识候选 / 发现（按归档确认）

| 关键词 | 一句话结论 | 出处 | 建议落点 | 类型 | 复利判断 | 处理结果 |
|--------|------------|------|----------|------|----------|----------|
