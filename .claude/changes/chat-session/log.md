# chat-session — Log

## 2026-04-26 — cc-propose

### 决策记录

1. **记忆策略**：选择滑动窗口 + LLM 摘要压缩。用户明确选择 MVP 即包含摘要能力。
   - 原因：对话训练场景中，上下文连续性对 AI 回复质量影响大，仅窗口会导致早期关键信息丢失
   - 摘要失败降级策略：降级为仅窗口，避免摘要失败阻断对话

2. **会话结束**：同时支持手动和自动结束
   - 自动结束：round_limit 到达时标记 completed
   - 手动结束：显式 API 调用
   - 两者不冲突，先触发的生效

3. **摘要不持久化**：MVP 阶段每轮重新生成摘要，保持简单
   - 原因：避免引入额外的存储设计和一致性维护
   - 后续可优化：持久化摘要到 session 表或新建摘要表

4. **HTTP 同步请求**：MVP 不做 WebSocket/Streaming
   - 原因：训练对话场景不需要实时流式输出，同步响应足够

5. **配置项单一**：仅新增 `memory_window_size`，LLM 相关配置复用已有 LLMConfig
   - 原因：摘要使用同一 LLM client，不单独配置

### 澄清与假设

- 用户确认：滑动窗口 + LLM 摘要（非仅窗口）
- 用户确认：手动 + 自动结束（非仅其中之一）
- 假设：round_limit=0 表示不限轮次
- 假设：摘要使用主 LLM client，不引入独立模型配置
- 假设：会话 ID 使用 UUID v4

### 触发的专题规则

- `rules/verification.md`：声明最低验证等级 L2，每个 task 映射验证 ID
- `rules/change-sizing.md`：分类为 M（4 个文件模块，一个主目标，清晰验证集群）

### 自动校验结果

- `cc-schema-check`：chat-session spec.md 和 tasks.md 无 schema 错误
- `cc-sync-check`：通过
- `cc-readset`：通过
- Harness 级别 pre-existing 失败：`docs/maintenance/`、`docs/adoption/`、`docs/examples/` 缺失（E_SCHEMA119），非本次变更引入

### cc-apply 执行记录

- T1: `internal/store/session.go` — Session/Message CRUD，build PASSED
- T2: `internal/memory/manager.go` — 滑动窗口+LLM摘要+降级，build+vet PASSED
- T3: `internal/session/service.go`, `errors.go` — 会话生命周期编排，build+vet PASSED
- T4: `internal/session/handler.go`, `config.go`, `main.go` — 4个HTTP端点+SessionConfig+接线，build+vet PASSED
- T5: 4个测试文件 — 22 tests passed (store 4, memory 4, session 14)

### 最终验证

- `go build ./...` — PASSED
- `go vet ./...` — PASSED
- `go test ./...` — 69 passed in 9 packages (22 new + 47 baseline)
- Baseline delta: 0 new failures
- spec.status → review

### 新增依赖

- `github.com/google/uuid` — 会话 ID 生成

### 未触发的专题规则

- `rules/database-changes.md`：使用现有表，不新增表或列
- `rules/api-compatibility.md`：全部为新增接口，无破坏性变更
- `rules/configuration.md`：仅新增 1 个配置项，已在 spec 记录
- `rules/security.md`：无资金、密钥、敏感数据变更
- `rules/observability.md`：日志观测点已在 spec 设计，但属于标准 INFO/ERROR 日志，无特殊需求

## 2026-04-27 — cc-review

### Stage 1: PASS — 所有 spec 需求项、业务规则、API 契约均有对应实现

### Stage 2: CONDITIONAL PASS — 2 Important + 3 Minor

| # | 严重度 | 问题 |
|---|--------|------|
| F1 | Important | Handler 直接访问 svc.store.CountMessages 和 json.Unmarshal，绕过 Service 层 |
| F2 | Important | Chat 中 3 次 json.Unmarshal 吞错误 + EndSession 中 CountMessages 吞错误 |
| F3 | Minor | contains/findSubstring 重复实现 strings.Contains |
| F4 | Minor | writeJSON 重复定义 |
| F5 | Minor | round_limit=0 无测试覆盖 |

## 2026-04-27 — cc-fix

### 修复记录

**F1 (Important) → fixed**

- 症状：Handler 直接访问 `h.svc.store.CountMessages()` 和 `json.Unmarshal(sess.RoleConfig)`
- 失败点：handler.go:157,190,192 绕过 Service 层
- 根因：Handler 需要聚合数据（currentRound/messageCount/roleDescription），但 Service 未提供对应方法
- 最小修复：新增 `GetSessionDetail` 返回聚合视图；`EndSession` 返回 `EndSessionResult`（含 FinalRound）；Handler 不再直接访问 store 或反序列化
- Guard：handler_test.go 中 TestHandleGet/TestHandleEnd 验证正确行为
- Fresh verification：23 tests passed

**F2 (Important) → fixed**

- 症状：3 次 `json.Unmarshal` 错误被静默忽略，1 次 `CountMessages` 错误被忽略
- 失败点：service.go:120-124,194
- 根因：Chat 中 Unmarshal 未做错误处理；EndSession 中 CountMessages 用 `_` 忽略
- 最小修复：新增 `parseSessionConfig` 方法集中处理 Unmarshal 并返回错误；EndSession 中 CountMessages 失败时记录 Warn 日志并降级为 0
- Guard：Unmarshal 错误现在会返回给调用方；EndSession 降级场景有 Warn 日志
- Fresh verification：23 tests passed

**F3 (Minor) → fixed**

- 替换自定义 `contains`/`findSubstring` 为 `strings.Contains`

**F4 (Minor) → accepted**

- `writeJSON` 重复定义在 role/handler.go 和 session/handler.go
- 接受理由：MVP 阶段两处实现完全相同且稳定，提取到共享包是合理的后续优化但不影响正确性

**F5 (Minor) → fixed**

- 新增 `TestChat_UnlimitedRounds`：验证 round_limit=0 时 isLast 始终为 false，会话保持 active

### 触发的专题规则

- `rules/verification.md`：修复后重新运行完整验证
- `rules/coding-style.md`：F2 修复后不再违反"禁止吞掉错误"
- `rules/debugging-workflow.md`：按 symptom → failure point → root cause → minimal fix → guard → verification 流程处理每个 Finding

### 未触发的专题规则

- 数据库（未修改 schema）、API 兼容性（无 breaking change）、安全（无敏感数据变更）、发布（MVP 新增模块）

### 最终验证

- `go build ./...` — PASSED
- `go vet ./...` — PASSED
- `go test ./...` — 70 passed in 9 packages

## 2026-04-27 — cc-archive

### 归档检查

- Review 状态：PASS（Stage 1 PASS, Stage 2 PASS）
- Open Findings：0（2 Important fixed, 2 Minor fixed, 1 Minor accepted）
- Blocked Tasks：0
- Unexplained Gaps：0
- Fresh Evidence：go build/vet/test 70 passed ✅

### 知识沉淀决策

- 沉淀文件：`knowledge/session-memory-patterns.md`（4 条：滑动窗口+摘要压缩、Service 聚合视图、集中 JSON Unmarshal、测试模式）
- 知识索引：新增 3 条踩坑记录（Handler 绕过 Service 层、集中 JSON 反序列化、摘要降级不阻断）
- 理由：会话记忆管理模式（窗口+摘要+降级）在后续分析引擎和多轮对话扩展中可复用；分层违规教训可防止其他 Handler 重复犯同样错误

### 归档验证

- `go build ./...` — PASSED
- `go vet ./...` — PASSED
- `go test ./...` — 70 passed in 9 packages
- `spec.status` → done
