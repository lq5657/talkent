---
change_id: e2e-integration
created: 2026-04-30
updated: 2026-04-30
---

### 任务拆分 — E2E 集成加固

#### 前置条件

* [ ] `spec.md` 已确认且 `status = apply`
* [ ] HARD-GATE 已通过
* [ ] 所有 Phase 0 change 已完成（scaffold-project, llm-client, role-and-goal, chat-session, analysis-engine, web-frontend）

#### Task 1: 服务级中间件

* **目标** : 实现 RequestID、Recovery、Timeout 三大中间件并在 server 中接线
* **不包含范围** : 不修改 handler 逻辑
* **涉及文件** :
  * `internal/server/middleware.go` — 新建，三个中间件
  * `internal/server/server.go` — 修改，New() 中链式包装中间件
  * `internal/config/config.go` — 修改，ServerConfig 新增 RequestTimeout
  * `config.example.yaml` — 修改，新增 request_timeout 配置项
* **关键签名** :
  ```go
  func RequestIDMiddleware(next http.Handler) http.Handler
  func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler
  func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler
  ```
  * `ServerConfig.RequestTimeout time.Duration` — 默认 30s
* **验收标准** : `curl -v` 可看到 `X-Request-ID` 响应头；超时请求返回 504；panic 返回 500 不崩溃；`go build ./...` 通过
* **验证步骤** : `go test ./internal/server/...` 新增中间件测试；手工 curl 验证三个场景
* **测试要求** : L3，chain 级验证
* **mapping_ids** : [V1]
* **依赖 / Wave** : wave-1，无前置依赖
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : 无（未配置 baseline 目录）

#### Task 2: 后端错误处理边界修复

* **目标** : 修复 analysis engine 日志脱敏 + handleGetReport 错误码区分
* **不包含范围** : 不修改其他 handler 错误处理，不新增前端代码
* **涉及文件** :
  * `internal/analysis/engine.go` — 修改，第 171 行日志脱敏
  * `internal/analysis/handler.go` — 修改，handleGetReport 区分 404/500
  * `internal/analysis/handler_test.go` — 修改，新增 404 测试用例
* **关键变更** :
  * `engine.go:171`：`"raw_output", truncated` → `"content_len", len(resp.Content)` + `"parse_error", parseErr.Error()`
  * `handler.go:88-93`：`GetLatestReport` 返回 `(nil, nil)` 时返回 `404` 而非 `500`
* **验收标准** : 分析 JSON 解析失败日志不含原始 LLM 输出；无报告 session 的 /report 返回 404；`go test ./internal/analysis/...` 通过
* **验证步骤** : `go test ./internal/analysis/... -v`
* **测试要求** : L2，package 级
* **mapping_ids** : [V2, V3]
* **依赖 / Wave** : wave-1，无前置依赖，可与 T1 并行
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : 无（未配置 baseline 目录）

#### Task 3: 前端错误处理与重试 UX

* **目标** : 前端网络错误感知、离线检测、ChatView 和 ReportView 重试按钮
* **不包含范围** : 不修改后端代码，不新增 Vue 路由或页面
* **涉及文件** :
  * `web/src/api/client.ts` — 修改，`apiRequest` 中区分 TypeError
  * `web/src/App.vue` — 修改，添加离线检测 banner
  * `web/src/views/ChatView.vue` — 修改，错误提示旁加重试按钮
  * `web/src/views/ReportView.vue` — 修改，分析失败加重试按钮
* **关键变更** :
  * `client.ts`：`catch` 块检测 `e instanceof TypeError`，消息改为"网络连接失败，请检查后端服务是否启动"
  * `App.vue`：`const onLine = ref(navigator.onLine)` + `window.addEventListener('online'/'offline', ...)`
  * `ChatView.vue`：错误 div 中添加 `<button @click="retryLastMessage">重试</button>`，从 `messages` 取最后一条 user 消息重新发送
  * `ReportView.vue`：错误 div 中添加 `<button @click="triggerAnalysis">重试</button>`
* **验收标准** : 后端未启动时访问前端不白屏，显示离线 banner；发送失败显示重试按钮，点击可重新发送；分析失败显示重试按钮
* **验证步骤** : 手工验证：停止后端 → 访问前端 → 确认 banner；恢复后端 → 重试成功
* **测试要求** : L4，manual 验证
* **mapping_ids** : [V4]
* **依赖 / Wave** : wave-1，无前置依赖，可与 T1/T2 并行
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : 无（未配置 baseline 目录）

#### Task 4: E2E 验证场景与手工回归

* **目标** : 建立 E2E 验证场景文档和 curl 脚本，执行全链路手工验证
* **不包含范围** : 不引入 Playwright/Cypress 自动化框架，不修改业务代码
* **涉及文件** :
  * `test/e2e/scenarios.md` — 新建，5 个 E2E 验证场景
  * `test/e2e/curl/` — 新建，curl 脚本集合
* **场景清单** :
  1. 正常全链路：设定角色 → 推荐目标 → 推荐维度 → 创建会话 → 多轮对话 → 结束 → 查看报告
  2. LLM 故障模拟：使用无效 API Key 启动服务，验证前端错误提示
  3. 空输入边界：空角色描述、空消息、空分析
  4. 边界轮数：round_limit=1 的极限场景
  5. 并发创建：同角色设定快速连续创建 3 个会话，数据正确隔离
* **验收标准** : 5 个场景逐一执行且结果记录在 `log.md`；curl 脚本集合可独立运行
* **验证步骤** : 手工执行 5 个场景，记录每个场景的关键结果
* **测试要求** : L4，manual 验证
* **mapping_ids** : [V5]
* **依赖 / Wave** : wave-2，依赖 T1/T2/T3 全部完成
* **回退方式** : 文档类，无需回退
* **完成后状态** : done
* **Baseline / Delta** : 无（未配置 baseline 目录）

#### Task 状态

| Task | 状态 | 依赖 | 并行安全 |
|------|------|------|----------|
| T1: 服务级中间件 | done | 无 | 可并行 (T2, T3) |
| T2: 后端错误处理边界修复 | done | 无 | 可并行 (T1, T3) |
| T3: 前端错误处理与重试 UX | done | 无 | 可并行 (T1, T2) |
| T4: E2E 验证场景与手工回归 | done | T1, T2, T3 | 否 |

```yaml
tasks:
  - name: 服务级中间件
    goal: 实现 RequestID、Recovery、Timeout 三大中间件并在 server 中接线
    files:
      - internal/server/middleware.go
      - internal/server/server.go
      - internal/config/config.go
      - config.example.yaml
    acceptance:
      - curl -v 看到 X-Request-ID 响应头
      - 超时请求返回 504
      - panic 返回 500 不崩溃
      - go build ./... 通过
    verification: go test ./internal/server/... + 手工 curl 验证
    test_requirement: L3 chain
    rollback: git revert
    baseline_delta: 无（未配置 baseline 目录）
    mapping_ids: [V1]
    status: done
  - name: 后端错误处理边界修复
    goal: 修复 analysis engine 日志脱敏 + handleGetReport 错误码区分
    files:
      - internal/analysis/engine.go
      - internal/analysis/handler.go
      - internal/analysis/handler_test.go
      - internal/analysis/engine_test.go
    acceptance:
      - 分析 JSON 解析失败日志不含原始 LLM 输出
      - 无报告 session 的 /report 返回 404
      - go test ./internal/analysis/... 通过
    verification: go test ./internal/analysis/... -v
    test_requirement: L2 package
    rollback: git revert
    baseline_delta: 无（未配置 baseline 目录）
    mapping_ids: [V2, V3]
    status: done
  - name: 前端错误处理与重试 UX
    goal: 前端网络错误感知、离线检测、ChatView 和 ReportView 重试按钮
    files:
      - web/src/api/client.ts
      - web/src/App.vue
      - web/src/views/ChatView.vue
      - web/src/views/ReportView.vue
    acceptance:
      - 后端未启动时访问前端不白屏，显示离线 banner
      - 发送失败显示重试按钮，点击可重新发送
      - 分析失败显示重试按钮
    verification: 手工验证：停止后端 → 访问前端 → 确认 banner → 恢复后端 → 重试成功
    test_requirement: L4 manual
    rollback: git revert
    baseline_delta: 无（未配置 baseline 目录）
    mapping_ids: [V4]
    status: done
  - name: E2E 验证场景与手工回归
    goal: 建立 E2E 验证场景文档和 curl 脚本，执行全链路手工验证
    files:
      - test/e2e/scenarios.md
      - test/e2e/curl/
    acceptance:
      - 5 个场景逐一执行且结果记录在 log.md
      - curl 脚本集合可独立运行
    verification: 手工执行 5 个场景，记录结果
    test_requirement: L4 manual
    rollback: 文档类无需回退
    baseline_delta: 无（未配置 baseline 目录）
    mapping_ids: [V5]
    status: done
```
