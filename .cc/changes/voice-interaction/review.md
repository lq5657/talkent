---
change_id: voice-interaction
reviewed_at: 2026-05-20 12:20
reviewer: Claude Code
stage1_status: pass
stage2_status: pass
final_status: pass
---

### Review Report — 流式文字 + 浏览器语音交互

#### 1. 输入材料

* `spec.md`：v1, 15 sections, HARD-GATE confirmed
* `tasks.md`：5 tasks, 2 waves
* `test-spec.md`：N/A (no test-spec created; per spec, tests verified via go test L2 + manual browser L4)
* `log.md`：5 commits, auto-validation evidence recorded
* 审查代码范围：14 files (internal/llm, internal/session, web/src/api, web/src/views, web/src/components)

#### 2. Task Coverage

| Task | 关联映射项 | 声明的验收标准 | 验证证据是否充分 | 闭环状态检查 | 结果 | 备注 |
|------|------------|----------------|------------------|--------------|------|------|
| Task 1: LLM ChatStream | V1 | go test ./internal/llm/ | 25 tests (4 new ChatStream) | ✓ 通过 | PASS | |
| Task 2: Session ChatStream | V2 | go test ./internal/session/ | 5 new ChatStream tests | ✓ 通过 | PASS | |
| Task 3: SSE Handler | V3 | curl + go test | 4 SSE handler tests | ✓ 通过 | PASS | L4 curl 验证待手工 |
| Task 4: Frontend Streaming | V4 | vue-tsc + vite build + 手工 | build passed | ✓ 通过 | PASS | L4 手工验证待执行 |
| Task 5: Frontend Voice UI | V4 | vue-tsc + vite build + 手工 | build passed | ✓ 通过 | PASS | L4 手工验证待执行 |

#### 2.1 验证映射检查

| 映射编号 | `spec.md` 声明状态 | 审查结论 | 证据 / 缺口 | 结果 |
|----------|--------------------|----------|-------------|------|
| V1 | L2 package test | ✓ | go test ./internal/llm/ — 25 passed, 4 ChatStream tests | PASS |
| V2 | L2 package test | ✓ | go test ./internal/session/ — 5 ChatStream tests | PASS |
| V3 | L2 package test | ✓ | go test ./internal/session/ handler — 4 SSE tests | PASS |
| V4 | L4 manual browser | ⚠️ | Build passes, but no manual browser verification evidence in log | GAP |

注：V4 的手工浏览器验证 (L4) 是 spec 明确要求的，目前只有构建成功证据。建议在 `cc-archive` 前补充手工验证证据或降级为 L4 build-only。

#### 2.2 Review Lens Matrix

| 镜头 | 触发原因 | 结论 | 是否形成 Finding |
|------|----------|------|------------------|
| spec-compliance | 默认 | PASS | 否 |
| verification-evidence | 默认 | V4 缺口 | 否（记录于 2.1） |
| robustness | 流式链路新架构 | 2 findings | 是 |
| performance | 未触发 | N/A | 否 |
| security | 未触发（浏览器 API，无 auth 变更） | N/A | 否 |
| api-contract | 新增 SSE endpoint | 1 finding | 是 |
| database-release | 未触发（无 DB schema 变更） | N/A | 否 |
| standards | JSON 构造方式 | 1 finding | 是 |

#### 3. Stage 1 — Spec Compliance

| # | 检查项 | 文件位置 | 结果 | 备注 |
|---|--------|----------|------|------|
| 1 | 缺失实现 | — | ✅ | F1-F4 全部实现，无缺失 |
| 2 | 多余实现 | handler.go chatResponse | ⚠️ | POST /chat 响应新增 user_message_created_at/assistant_message_created_at 字段未在 spec 接口变更表中声明（compatible_addition，无 breaking impact） |
| 3 | 理解偏差 | — | ✅ | 无偏差 |
| 4 | 业务规则落地 | — | ✅ | 全部 7 条规则已落地（feature detection、自动 TTS、toggle play/stop、SSE 断线 fallback、session 校验复用） |
| 5 | 对外契约准确性 | handler.go:24 | ✅ | SSE endpoint 路径、headers、payload 格式均与 spec 一致 |

#### 4. Stage 2 — Code Quality

| 级别 | 问题类型 | 文件位置 | 结果 | 建议 |
|------|----------|----------|------|------|
| Critical | 安全/资金/并发/数据丢失 | — | ✅ | 无 |
| Important | 错误/context/校验/魔法值/兼容风险 | handler.go:242; ChatView.vue:108-115 | ⚠️ | 见 Findings F1, F2 |
| Minor | 文档/注释/import | service.go:257 | ⚠️ | 见 Finding F3 |

#### 5. Findings

| # | 级别 | 描述 | 位置 | 建议动作 | 状态 |
|---|------|------|------|----------|------|
| F1 | Important | SSE error event 使用 `%s` 直接嵌入 JSON — error message 中若含 `"`、`\` 或换行符会产生非法 JSON | handler.go:242 `fmt.Fprintf(w, "data: {\"error\":\"%s\"}\n\n", chunk.Error.Error())` | 改为 `%q` 或 `json.Marshal` 转义（与 token/done 事件保持一致——它们已使用 `%q`） | open |
| F2 | Important | SSE 流式部分成功时 fallback 产生重复消息 — `aiMsg` 已有部分 token 时 `messages.value.pop()` 不执行，但 fallback 的 `api.chat()` 会追加第二条 AI 消息 | ChatView.vue:108-115 | 在 fallback 前始终移除部分流式消息，或用 fallback 响应替换其 content | open |
| F3 | Minor | Service goroutine 在客户端断开时可能阻塞 — `ChatStream` 的转发 goroutine 不检查 context cancellation，依赖 LLM stream 传播 ctx 错误；若 channel buffer (cap=8) 先填满，goroutine 会永久阻塞 | service.go:258 | 在 goroutine 内增加 `select` 监听 `ctx.Done()`，或在 `ch <-` 前检查 context | open |
| F4 | Minor | POST /chat 响应新增字段未记录于 spec — `user_message_created_at` 和 `assistant_message_created_at` 是新字段，属于 compatible_addition 但未在 spec §6 接口变更表中列出 | handler.go:86-87 (chatResponse); spec.md §6 | 在 spec §6 中补充新增字段说明，或注明该字段变更属于已归档的 message-timing change | open |

#### 5.1 Accepted Findings 确认记录（按需）

（待用户选择后填写）

#### 6. 结论

* **Stage 1 结论**：**PASS** — spec 全覆盖，无缺失功能，无理解偏差，业务规则全部落地
* **Stage 2 结论**：**PASS** — 2 个 Important findings (F1, F2 建议在 `cc-fix` 中修复), 2 个 Minor findings (F3, F4 可接受或修复)
* **总体结论**：可进入 `cc-fix` 修复 F1-F2，或直接进入 `cc-archive`（接受 F1-F4 剩余风险）
