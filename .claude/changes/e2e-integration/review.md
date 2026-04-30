---
change_id: e2e-integration
stage1_status: partial
stage2_status: skipped
final_status: partial
findings:
  - level: Critical
    description: "中间件顺序错误导致 panic recovery 失效。RecoveryMiddleware 在 TimeoutMiddleware 外层，但 TimeoutMiddleware 通过 go func() 在独立 goroutine 中执行下游 handler。当 handler panic 时，panic 发生在子 goroutine，RecoveryMiddleware 的 defer recover() 在父 goroutine 无法捕获，进程直接崩溃。违反 spec 验收标准'panic 返回 500 不崩溃'。修复：交换 Recovery 和 Timeout 的顺序。"
    status: fixed
    location: internal/server/server.go:31-34
  - level: Important
    description: "tasks.md T1 RecoveryMiddleware 签名不匹配。task 声明 func RecoveryMiddleware(next http.Handler) http.Handler，实际实现是工厂函数 func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler。"
    status: fixed
    location: .claude/changes/e2e-integration/tasks.md:T1
  - level: Important
    description: "tasks.md T2 声明 engine_test.go 验证日志不含 raw_output，但现有测试未捕获或断言日志输出。代码修改正确，但测试验证声明缺乏对应实现。"
    status: fixed
    location: .claude/changes/e2e-integration/tasks.md:T2
  - level: Important
    description: "tasks.md T1 '不包含范围'声明矛盾：不新增中间件测试文件 vs 验证步骤要求新增中间件测试。实际已正确创建 middleware_test.go，应修正文档。"
    status: fixed
    location: .claude/changes/e2e-integration/tasks.md:T1
  - level: Important
    description: "验证映射 V1-V5 全部 todo，未在 cc-apply 中同步为 apply-covered。V1-V3 已同步为 apply-covered，V4/V5 为 manual 验证待执行。"
    status: fixed
    location: .claude/changes/e2e-integration/spec.md:§8
  - level: Important
    description: "T4 执行记录表为空，log.md 中无 5 个场景的实际执行证据。cc-archive 前必须补充。"
    status: open
    location: test/e2e/scenarios.md:221-227
  - level: Minor
    description: "baseline/ 目录未创建，但 tasks 中声明了 baseline delta。已从 tasks.md 移除所有 baseline delta 声明。"
    status: fixed
    location: .claude/changes/e2e-integration/tasks.md:T1-T4
  - level: Minor
    description: "Scenario 2 无对应 curl 脚本（预期为手工浏览器验证），已在 run-all.sh 中增加说明。"
    status: fixed
    location: test/e2e/curl/
---

# Review Report — e2e-integration

## 审查概要

- **Stage 1 (spec compliance)**: FAILED — 1 Critical finding
- **Stage 2 (code quality)**: SKIPPED — blocked by Stage 1 Critical
- **最终结论**: **FAIL** — C1 必须修复后才能进入 Stage 2 和归档

---

## Findings Detail

### Critical

**C1 — `internal/server/server.go:31-34`**: 中间件顺序错误导致 panic recovery 失效。

`RecoveryMiddleware` 在 `TimeoutMiddleware` 外层，但 `TimeoutMiddleware` 通过 `go func()` 在独立 goroutine 中执行下游 handler。当 handler panic 时，panic 发生在子 goroutine，`RecoveryMiddleware` 的 `defer recover()` 在父 goroutine 无法捕获，进程直接崩溃。

违反 spec 验收标准"panic 返回 500 不崩溃"。

**根因分析**：

```
当前（错误）: RequestID → Recovery → Timeout(goroutine: CORS → Handler)
Recovery 在父goroutine，捕获不到子goroutine的panic

修复后: RequestID → Timeout(goroutine: Recovery → CORS → Handler)
Recovery 在子goroutine内，能捕获handler/CORS的panic
```

**修复**（`server.go:30-34`）：
```go
var handler http.Handler = mux
handler = corsMiddleware(handler)
handler = RecoveryMiddleware(logger)(handler)    // 移到 Timeout 内层
handler = TimeoutMiddleware(timeout)(handler)
handler = RequestIDMiddleware(handler)
```

同步更新 spec.md §3 feature 3 中的中间件顺序描述。

### Important

**I1 — `tasks.md:T1`**: `RecoveryMiddleware` 签名不匹配。task 声明 `func RecoveryMiddleware(next http.Handler) http.Handler`，实际实现是 `func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler`（工厂函数）。task 文档需更新为实际签名。

**I2 — `tasks.md:T2`**: Task 声明 `engine_test.go` 修改为"验证日志不含 raw_output"，但现有测试未捕获或断言日志输出内容。代码修改正确（`raw_output` → `content_len`），但测试验证声明缺乏对应实现。需补充日志捕获测试或更新 task 说明。

**I3 — `tasks.md:T1`**: "不包含范围"声明矛盾："不新增中间件测试文件" vs "验证步骤: go test 新增中间件测试"。实际已正确创建 `middleware_test.go`，应修正 task 的"不包含范围"。

**I4 — `spec.md:§8`**: 验证映射状态 V1-V5 全部 `todo`，未在 `cc-apply` 中同步为 `apply-covered`。违反 `verification.md` 的映射同步要求。

**I5 — `test/e2e/scenarios.md:221-227`**: T4 执行记录表为空，`log.md` 中无 5 个场景的实际执行证据。V5 验证状态为 `todo`。`cc-archive` 前必须补充。

### Minor

**M1 — `tasks.md:T1-T4`**: `baseline/` 目录未创建，但 tasks 中声明了 baseline delta。需创建或移除声明。

**M2 — `test/e2e/curl/`**: Scenario 2 无对应 curl 脚本（预期为手工浏览器验证），建议在 `run-all.sh` 中增加说明。

---

## Task Coverage Matrix

| Task | Status | 文件存在 | 验收标准 | 验证证据 | 结论 |
|------|--------|----------|----------|----------|------|
| T1 | done | YES | 部分达标 | 26 tests pass | C1 阻塞 — panic recovery 在集成场景失效 |
| T2 | done | YES | 达标 | 20 tests pass | 代码正确，I2 文档不一致 |
| T3 | done | YES | 达标 | `npm run build` pass | 实现正确 |
| T4 | done | YES | **未达标** | 无执行证据 | I5 — 场景未实际执行 |

---

## Validation Mapping

| ID | 描述 | 声明等级 | 实际证据 | 状态 |
|----|------|----------|----------|------|
| V1 | 中间件功能正确 | L3 chain | 6 tests pass (isolated) + C1 集成缺陷 | gap |
| V2 | 日志不含原始LLM输出 | L2 package | 代码变更正确，日志断言测试缺失 | apply-covered |
| V3 | 错误码区分 | L2 package | handleGetReport 404/500 区分已实现 | apply-covered |
| V4 | 前端错误处理 | L4 manual | build pass，未手工验证 | todo |
| V5 | E2E全链路验证 | L4 manual | 脚本已创建，未执行 | todo |

---

## 专题规则命中

| 规则 | 命中原因 | 检查结果 |
|------|----------|----------|
| `security.md` | engine.go 日志脱敏 | PASS — LLM 原始输出已移除 |
| `coding-style.md` | 新增 Go 代码 | PASS — 命名、错误处理、日志规范符合要求 |
| `verification.md` | 验证等级与映射 | FAIL — 映射未同步，T4 无执行证据 |
| `api-compatibility.md` | 内部接口变更 | PASS — 纯新增，无 breaking change |
| `configuration.md` | 新增 `request_timeout` | PASS — 有默认值，已记录 |
| `testing-strategy.md` | 测试分层 | PASS — L2/L3 分层合理 |

---

## 下一步

1. **cc-fix e2e-integration** — 修复 C1（中间件顺序），同步修复 I1-I5 文档不一致
2. C1 修复后重新运行 `cc-review` 进入 Stage 2（code quality）
3. T4 场景手工执行并获得 V4/V5 fresh evidence
4. 全部通过后 `cc-archive e2e-integration`
