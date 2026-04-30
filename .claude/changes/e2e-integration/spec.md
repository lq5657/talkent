---
change_id: e2e-integration
status: review
depends_on:
  - scaffold-project
  - llm-client
  - role-and-goal
  - chat-session
  - analysis-engine
  - web-frontend
parallel_safe: false
branch: feat/e2e-integration
created: 2026-04-30
updated: 2026-04-30
complexity: medium
min_validation_level: L3
---

### E2E 集成加固

#### 0.1 需求收敛记录

`cc-new-project 项目定义` → `MVP 路线图 Phase 1` → `E2E 集成：全链路联调、错误处理、边界覆盖`（收敛方式：backlog）

#### 1. 背景与目标

Phase 0 的 6 个 change 完成了所有功能模块（scaffold → llm-client → role-and-goal → chat-session → analysis-engine → web-frontend）。但以下横切关注点缺失：

- **服务级中间件**：无请求超时、panic recovery、请求 ID
- **错误处理边界缺口**：analysis engine 日志记录原始 LLM 输出（安全隐患）、handleGetReport 不区分 404/500
- **前端网络异常处理**：`fetch` 网络错误给出泛化消息，无重试/离线提示
- **E2E 验证**：无全链路验证场景，无法确认真实环境可用性

本次 change 补齐以上四项，使系统达到生产就绪度的最低基线。

#### 1.1 本次不做

- 引入新路由库（chi/gin）
- OpenTelemetry / Prometheus 等可观测性平台集成
- 数据库 schema 变更或 migration
- 前端框架升级或 UI 重构
- CI/CD 流水线搭建
- Playwright / Cypress 自动化 E2E 测试框架

#### 2. 代码现状（Research Findings）

- 后端所有模块（config/llm/log/server/store/role/session/memory/analysis）已实现，95 tests passing
- 中间件现状：仅 `corsMiddleware`，无超时/恢复/请求ID
- 错误处理：handler 层有 sentinel error 分类（`ErrSessionNotFound`、`ErrSessionCompleted`、`ErrSessionNotCompleted`），但 `handleGetReport` 错误码区分不完整
- `analysis/engine.go:171`：JSON 解析失败重试后仍失败时，`logger.Error` 包含 `raw_output` 字段（最多 500 字符原始 LLM 输出）
- 前端 `client.ts`：`apiRequest` 中 `fetch` 的 `TypeError` 与 HTTP 错误混在一起
- 前端三 View：均有 try/catch 错误展示，但无重试按钮和离线检测
- 配置现状：`ServerConfig` 仅有 `Host`/`Port`，无 `RequestTimeout`

#### 3. 功能点

* [ ] **服务中间件**：`internal/server/middleware.go`，RequestID + Recovery + Timeout 三个中间件
* [ ] **配置新增**：`ServerConfig.RequestTimeout`（默认 30s），`config.example.yaml` 同步
* [ ] **中间件接线**：`server.go` 中按 RequestID → Recovery → Timeout → CORS 顺序链式包装
* [ ] **analysis 日志脱敏**：`engine.go:171` 移除 raw_output，改为 content_len + parse_error
* [ ] **handleGetReport 错误码修正**：区分 404（无报告）和 500（查询失败）
* [ ] **前端网络错误感知**：`client.ts` 区分 TypeError 和 HTTP 错误
* [ ] **前端离线检测**：`App.vue` 监听 online/offline 事件，离线时显示 banner
* [ ] **前端重试按钮**：ChatView 和 ReportView 错误提示旁增加"重试"按钮
* [ ] **E2E 验证场景**：`test/e2e/scenarios.md` + `test/e2e/curl/` 脚本集合

#### 4. 技术决策

| 决策点 | 推荐方案 | 拒绝方案 | 理由 |
|--------|----------|----------|------|
| 中间件实现 | 标准 `func(http.Handler) http.Handler` | 引入 chi/gin 中间件 | 不引入新路由框架依赖，保持 `net/http` 一致性 |
| RequestTimeout 默认值 | 30s | 10s / 60s | 与现有 LLM timeout（30s）一致，分析请求可用 |
| 分析日志脱敏 | 记录 content_len + parse_error | 完全移除日志 | 保留错误定位能力但移除敏感输出 |
| 前端离线检测 | `navigator.onLine` + 事件监听 | WebSocket 心跳 | 简单可靠，无需后端配合 |
| E2E 验证形式 | 手工场景 + curl 脚本 | Playwright/Cypress | 先建立基线，自动化框架在后续 change 引入 |

#### 5. 兼容性

| 维度 | 分类 | 说明 |
|------|------|------|
| API 接口 | compatible_addition | 无破坏性变更，新增响应头 `X-Request-ID` |
| 配置 | compatible_addition | 新增可选 `server.request_timeout`，有默认值 30s |
| 数据库 | 不变更 | 无 schema / migration |
| 前端 | 增强 | 不改变现有交互流程，增量添加重试/离线能力 |

#### 6. 发布与回滚

- 发布方式：单服务替换重启
- 回滚路径：代码 revert + 重新构建
- 新增配置 `server.request_timeout` 有安全默认值，旧配置文件无需修改
- 观察窗口：启动后执行 §9 全链路验证场景，确认 5 个场景通过

#### 7. 风险

| 风险 | 等级 | 缓解 |
|------|------|------|
| RequestTimeout 截断分析请求 | 低 | 默认 30s，分析请求使用独立 LLM timeout（30s），超时可单独调大 |
| panic recovery 掩盖编程错误 | 低 | 记录完整 stack trace，不静默吞错 |
| 前端重试产生重复消息 | 低 | Chat 接口非幂等；重试仅重新发送最后一条用户消息，session 级去重 |

#### 8. 验证映射 (validation_map)

| ID | 验证项 | Level | Evidence Type | 描述 | Task | Status |
|----|--------|-------|---------------|------|------|--------|
| V1 | 中间件链：request_id 注入、超时 504、panic 500 | L3 | chain | curl 验证 X-Request-ID 响应头 + 超时返回 504 + panic 返回 500 | T1 | todo |
| V2 | analysis engine 错误日志不含原始 LLM 输出 | L2 | package | 验证 JSON 解析失败重试后日志仅含 content_len 和 parse_error | T2 | todo |
| V3 | handleGetReport 对无报告 session 返回 404 | L2 | package | httptest 验证 404 vs 500 区分 | T2 | todo |
| V4 | 前端网络错误显示离线提示 + 重试按钮可用 | L4 | manual | 停止后端 → 前端不崩溃显示离线 banner；恢复后端 → 重试成功 | T3 | todo |
| V5 | 全链路手工验证通过 | L4 | manual | 设定→对话→分析→报告 完整流程无阻断 | T4 | todo |

#### 9. HARD-GATE

* **confirmed_at**：2026-04-30
* **confirmed_by**：lq5657
* **confirmed_spec_revision**：`spec.md` @ 2026-04-30，9 个功能点，V1-V5 验证映射（7 列），中间件+错误处理修复+前端UX+E2E验证
* **confirmed_tasks_revision**：`tasks.md` @ 2026-04-30，4 个 task，T1-T3 可并行
* **confirmed_scope**：`internal/server/middleware.go` 三个中间件；`internal/config/config.go` ServerConfig.RequestTimeout；`internal/analysis/engine.go` 日志脱敏；`internal/analysis/handler.go` 404/500 区分；`web/src/api/client.ts` 网络错误感知；`web/src/App.vue` 离线检测；`web/src/views/ChatView.vue` 重试；`web/src/views/ReportView.vue` 重试；`test/e2e/` E2E 验证场景；不含路由库升级/可观测性平台/Schema 变更/自动化 E2E 框架
* **accepted_risks**：RequestTimeout 对分析请求可能偏紧（默认 30s 足够）；panic recovery 需确保 stack trace 完整记录；Chat 重试非幂等（最后一条消息重复发送风险低）
* **human_review_required**：false
* **human_review_status**：not_required
