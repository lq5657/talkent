---
change_id: analysis-engine
reviewed_at: 2026-04-28
reviewed_by: cc-review
stage_1: pass_with_findings
stage_2: pass_with_findings
verdict: pass
fix_round: 1
fixed_at: 2026-04-29
fixed_by: cc-fix
---

### Stage 1: Spec Compliance Review

| 需求项 | 规格位置 | 实现状态 | 判定 | 备注 |
|--------|----------|----------|------|------|
| Analysis Store CRUD | §3-1 | `internal/store/analysis.go` CreateReport/GetLatestReport/ListReports | ✅ pass | |
| Analysis Engine | §3-2 | `internal/analysis/engine.go` buildPrompt/callWithRetry/parseDimensionResults/renderMarkdown | ✅ pass | |
| Analysis Service | §3-3 | `internal/analysis/service.go` TriggerAnalysis/GetLatestReport/ListReports + Analyzer 接口 | ✅ pass | |
| 手动触发 API | §3-4 | `POST /api/sessions/{id}/analyze` → 201 | ✅ pass | |
| 自动触发钩子 | §3-5 | OnSessionEndFunc + main.go 回调注入 | ✅ pass | F1 fixed |
| 报告查询 API | §3-6 | GET /report + GET /reports | ✅ pass | |
| Schema 迁移 | §3-7 | schema.go DDL + db.go runMigrations | ✅ pass | F4 fixed |
| 配置项 | §3-8 | AnalysisConfig.AutoTrigger + env var | ✅ pass | |
| main.go 接线 | §3-9 | 完整依赖链 + 路由注册 | ✅ pass | |
| 仅 completed 可分析 | §4 | service.go:42-44, handler 409 | ✅ pass | |
| 维度来自会话 dimensions | §4 | service.go:59-61 | ✅ pass | |
| 一次调用全维度 | §4 | engine.go 单次 Chat | ✅ pass | |
| JSON 解析失败重试一次 | §4 | callWithRetry | ✅ pass | |
| 重试失败截断 500 字符 | §4 | engine.go:171-176 | ✅ pass | |
| 允许重复分析 | §4 | 每次独立 CreateReport | ✅ pass | |
| JSON + Markdown 共存 | §4 | AnalysisResult 两个字段 | ✅ pass | |
| model_used 记录 | §4 | engine.go:61, service.go:78 | ✅ pass | |
| System/User Prompt 分离 | §4.2 | engine.go:111-114, 测试覆盖 | ✅ pass | |
| 可观测性：分析触发 | §8 | service.go:46 | ✅ pass | |
| 可观测性：LLM 调用 | §8 | engine.go:49 | ✅ pass | F5 fixed |
| 可观测性：JSON 失败 WARN | §8 | engine.go:155 | ✅ pass | |
| 可观测性：重试失败 ERROR | §8 | engine.go:175 | ✅ pass | |
| 可观测性：报告保存 | §8 | service.go:85 | ✅ pass | |
| 可观测性：自动触发失败 WARN | §8 | main.go:66 | ✅ pass | |

### Stage 2: Code Quality Review

#### Findings

| ID | 严重度 | 位置 | 描述 | 状态 |
|----|--------|------|------|------|
| F1 | Critical | `internal/session/service.go:176-182` | Chat 自动结束路径未调用 `notifySessionEnd`，导致会话达到轮次限制自动结束时不会触发分析回调 | fixed |
| F2 | Minor | `internal/analysis/engine.go:147-181` | `callWithRetry` 内部解析后 `Analyze` 再次解析，冗余计算 | fixed |
| F3 | Minor | `internal/analysis/handler.go:138` | `reportToResponse` 中 `json.Unmarshal` 错误被静默忽略 | fixed |
| F4 | Minor | `internal/store/db.go:34-43` | `runMigrations` 对所有 ALTER TABLE 错误一律静默忽略 | fixed |
| F5 | Minor | `internal/analysis/engine.go:49` | Engine 日志缺 session_id 字段，不满足 spec §8 可观测性要求 | fixed |

#### Fix 证据

| ID | 修复方式 | 涉及文件 | Guard | Fresh Evidence |
|----|----------|----------|-------|----------------|
| F1 | Chat isLast 分支添加 `s.notifySessionEnd(sessionID)` | `internal/session/service.go` | `TestChat_AutoEndNotifiesOnSessionEnd` | 95 tests pass |
| F2 | `callWithRetry` 返回 `([]DimensionResult, string, error)`，`Analyze` 不再重复解析 | `internal/analysis/engine.go` | 现有测试仍通过 | 95 tests pass |
| F3 | `reportToResponse` 中 `json.Unmarshal` 错误时设 `dims = nil` | `internal/analysis/handler.go` | 现有测试仍通过 | 95 tests pass |
| F4 | `runMigrations` 增加 `isDuplicateColumnError` 检查，非"列已存在"错误向上返回 | `internal/store/db.go` | 现有测试仍通过 | 95 tests pass |
| F5 | `Analyze` 新增 `sessionID` 参数，日志补充 `session_id` 字段；`Analyzer` 接口同步更新 | `internal/analysis/engine.go`, `service.go`, `engine_test.go`, `service_test.go` | 现有测试适配新签名 | 95 tests pass |

### 验证映射检查

| 编号 | 验证项 | 状态 | 证据 | 备注 |
|------|--------|------|------|------|
| V1 | Analysis Store CRUD | apply-covered | analysis_test.go | ✅ |
| V2 | 分析 Prompt 正确构造 | apply-covered | engine_test.go | ✅ |
| V3 | LLM 一次调用全维度分析 | apply-covered | engine_test.go | ✅ |
| V4 | JSON 解析失败重试 | apply-covered | engine_test.go | ✅ |
| V5 | 重试仍失败返回错误 | apply-covered | engine_test.go | ✅ |
| V6 | Markdown 报告正确渲染 | apply-covered | engine_test.go | ✅ |
| V7 | 仅 completed 会话可分析 | apply-covered | service_test.go + handler_test.go | ✅ |
| V8 | 手动触发 API | apply-covered | handler_test.go | ✅ |
| V9 | 自动触发钩子 | apply-covered | service_test.go + session/service_test.go (F1 guard) | ✅ F1 修复后闭环 |
| V10 | 报告查询 API | apply-covered | handler_test.go | ✅ |
| V11 | Schema 迁移正确执行 | apply-covered | analysis_test.go | ✅ |
| V12 | main.go 接线后服务启动正常 | apply-covered | go build + TestHandler_RoutesRegistered | ✅ |

### 风险镜头

| 规则 | 是否命中 | 说明 |
|------|----------|------|
| database-changes | ✅ | Schema 迁移为 expand 类型，新增列有默认值；F4 修复后非预期错误可上报 |
| api-compatibility | ✅ | 3 个端点均为 compatible_addition；F5 修改 Analyzer 接口为内部接口，不影响对外 API |
| security | ✅ | System/User Prompt 分离已验证；无资金/权限高风险点；LLM 原始输出截断保护 |
| configuration | ✅ | AnalysisConfig 声明完整，默认值安全 |
| observability | ✅ | F5 修复后所有 spec §8 日志点完整覆盖 |
| testing-strategy | ✅ | L2 验证等级与风险匹配；95 tests pass |

### 综合结论

**verdict: pass**

所有 5 个 Finding（1 Critical + 4 Minor）均已修复，fresh evidence 为 `go build ./...` 成功、`go test ./...` 95 passed、`go vet` clean。V1-V12 全部闭环，spec 与实现一致。

无 open Finding，无阻塞项，可以归档。

### 建议下一步

- `cc-archive analysis-engine`
