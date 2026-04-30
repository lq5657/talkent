### 知识索引

领域知识的轻量索引。每条用一句话说清核心逻辑。
格式：`**触发关键词** : 一句话核心逻辑 → 文件路径.函数名（可选）`

---

### 触发条件（何时向用户建议沉淀）

满足以下任一条件时，应建议沉淀到 knowledge/：

| 条件 | 示例 |
|------|------|
| 发现 Go 最佳实践与团队规范的差异 | "发现 xx 场景应该用 xxx 模式" |
| 踩坑并找到解决方案 | 某问题排查过程有价值 |
| 技术决策及其理由 | 选了 A 方案而未选 B |
| 业务规则的特殊处理 | 某字段在 xx 情况下需要特殊处理 |
| 第三方库的坑 | 某库在 xx 场景下有 bug |

---

### 知识分类

#### 业务知识

业务领域特定规则，通常与数据模型和状态机相关。

#### 技术约定

项目特定的实现模式、框架使用规范。

#### 踩坑记录

问题描述 + 原因 + 解决方案。可被 future self 快速查找复用。

---

### 沉淀流程

1. 发现有价值的信息
2. 用 `/** 关键词 */` 格式标注在发现位置
3. 在对应分类下追加条目
4. 格式：`**关键词** : 核心逻辑 → 出处`

---

#### 业务知识

（随实践积累补充）

#### 技术约定

* **Goroutine并发退出** : 使用 `golang.org/x/sync/errgroup` 管理并发退出，context 传播取消信号，禁止手动 `sync.WaitGroup` 除非有特殊原因
* **触发场景** : 新增涉及并发 HTTP 调用/gRPC 调用/goroutine 启动的代码时
* **errgroup 典型模式** :
  ```go
  g, ctx := errgroup.WithContext(ctx)
  g.Go(func() error {
      return doSomething(ctx)
  })
  if err := g.Wait(); err != nil {
      // 错误处理
  }
  ```
* **Harness试点验收** : 发布前先用验收清单确认生命周期、规则边界、reviewer 契约和样例是否可跑通 → `docs/adoption/pilot-checklist.md`
* **Harness接入预检** : 当需要验收或排查当前项目对框架的接入完整性时，执行 `cc-preflight`；`docs/adoption/integration-preflight-checklist.md` 是该命令的执行依据，而不是隐性入口 → `docs/maintenance/legacy/commands/cc-preflight.md`
* **Harness长期记忆** : 项目导航写 `context/dev-map.md`，change 看板写 `changes/task-board.md`，写入边界遵守 `rules/memory-policy.md`，角色权限遵守 `rules/role-contracts.md`
* **cc-new-project回归评测** : 修改 `cc-new-project`、项目级模板或项目生命周期路由后，先用回归样例验证是否仍能正确完成项目定义、MVP 路线图和首批 change backlog，并自然桥接到后续 `cc-propose` → `docs/maintenance/cc-new-project-eval-cases.md`
* **Harness接入高频问题** : 接入真实项目时，优先排查 `cc-init` 边界、路径解释、命令冲突、checkpoint 展示契约和验证等级等常见跑偏点 → `docs/maintenance/common-integration-pitfalls.md`
* **cc-propose回归评测** : 修改 `cc-propose` 提问策略、路由边界、roadmap 对齐逻辑或 checkpoint 后，先用回归样例验证是否仍能正确区分新项目与已有项目 change，并把 change 放回 phase / backlog 语义中收敛 → `docs/maintenance/cc-propose-eval-cases.md`
* **cc-apply回归评测** : 修改 `cc-apply`、验证等级规则、测试策略、task 顺序约束或相关样例后，先用回归样例验证不会退化成“只保 `go build`”，也不会跳过 `依赖 / Wave` 强行推进 → `docs/maintenance/cc-apply-eval-cases.md`
* **Harness协议回归评测** : 修改命令口径、机器工作流、生命周期状态机、HARD-GATE、Git 策略、验证矩阵、schema 或校验脚本后，先跑协议回归样例与 `cc-verify` → `docs/maintenance/cc-harness-protocol-eval-cases.md`
* **cc-review回归重点** : 审查实现结果时，除了核对 spec 和证据，还要检查 task promised outcome、roadmap 对齐和执行顺序是否真正落地 → `docs/maintenance/legacy/commands/cc-review.md`
* **系统讲解命令** : 当目标是帮助用户深入掌握大型复杂项目时，使用 `cc-explain-system` 输出系统定位、架构、数据流、技术机制、难点与阅读路径 → `commands/cc-explain-system.md`
* **Checkpoint结果列** : 所有 checkpoint 表的状态必须写入 `结果` 列，禁止把 `[x]` / `[ ]` 塞进 `检查项` 列冒充结果，结果值只允许 `✅`、`❌`、`⚠️`、`N/A`
* **具体优于抽象** : 默认先写具体实现，只有在多实现、测试替身或明确解耦诉求出现时再抽象接口
* **接口后置设计** : 接口从调用方需要出发定义，优先小接口，避免为了“未来可能需要”预留
* **无数据不优化** : 性能优化先看 benchmark、pprof、逃逸分析，没有数据就不要引入 `sync.Pool` 或复杂优化
* **Channel不是默认更优** : `channel` 适合表达数据流，`mutex/atomic` 适合本地共享状态，需按场景选择同步策略

#### 踩坑记录

* **模板匹配优先LLM兜底** : 推荐链路用二级策略（查表/模板匹配→LLM 生成），返回值带 source 字段区分来源，System Prompt 硬编码防注入 → `knowledge/role-patterns.md`
* **server.New回调路由注册** : `registerRoutes func(*http.ServeMux)` 回调模式，各 Handler 实现 RegisterRoutes 方法，新增模块无需改 server.go → `knowledge/role-patterns.md`
* **capturingClient测试模式** : 包装 mockLLMClient 拦截 ChatMessage 切片，验证 System Prompt 不被用户输入污染，与 llm-client 的 httptest.NewServer 模式互补 → `knowledge/role-patterns.md`
* **Handler绕过Service层** : Handler 直接访问 svc.store 或做 json.Unmarshal 是分层违规，应通过 Service 提供的聚合视图方法（如 GetSessionDetail）获取数据 → `knowledge/session-memory-patterns.md`
* **集中JSON反序列化** : store 中多个 JSON 字段需要反序列化时，集中到单一方法处理并检查错误，避免静默零值传入业务逻辑 → `knowledge/session-memory-patterns.md`
* **摘要降级不阻断** : LLM 摘要失败时降级为仅窗口模式，memory_source 返回 "window"，不阻断对话主流程 → `knowledge/session-memory-patterns.md`
* **OnSessionEndFunc回调解耦** : session.Service 通过回调字段暴露钩子，main.go 注入具体实现，避免循环依赖；Chat 自动结束和 EndSession 都必须调用 notifySessionEnd → `knowledge/analysis-patterns.md`
* **SQLite migration双轨** : CREATE TABLE 包含所有列 + ALTER TABLE for 旧库 + isDuplicateColumnError 区分可忽略错误与真实失败；新列必须有 NOT NULL DEFAULT → `knowledge/analysis-patterns.md`
* **LLM JSON容错** : 结构化 Prompt + code fence stripping + callWithRetry（一次重试 + 温度降低 + 截断日志） + 直接返回解析结果避免冗余 → `knowledge/llm-client-patterns.md`
* **报告双重格式** : analysis_reports 同时存 dimension_results(JSON) + markdown_content(Markdown)，Handler 反序列化错误时设 nil 而非静默吞错 → `knowledge/analysis-patterns.md`
* **Tailwind v4 Vite集成** : v4 无需 tailwind.config.js，使用 @tailwindcss/vite 插件 + `@import "tailwindcss"` → `knowledge/web-frontend-patterns.md`
* **marked v15代码高亮** : marked.use({ renderer: { code } }) 替代已移除的 setOptions({ highlight }) → `knowledge/web-frontend-patterns.md`
* **v-html XSS防护** : marked v15 无内置 sanitize，DOMPurify.sanitize() 是 v-html 渲染 Markdown 的强制红线 → `knowledge/web-frontend-patterns.md`
* **CORS中间件最佳实践** : Allow-Methods/Headers 仅对允许源设置 + Max-Age:86400 减少预检 + Vite 代理免 CORS → `knowledge/web-frontend-patterns.md`
* **highlight.js分包** : manualChunks 拆分 970KB 包为按需加载 chunk，避免首屏影响 → `knowledge/web-frontend-patterns.md`
* **中间件Goroutine安全** : RecoveryMiddleware 必须在 TimeoutMiddleware 内层（子 goroutine 内），否则子 goroutine panic 会崩溃进程 → `knowledge/e2e-integration-patterns.md`
* **LLM日志脱敏** : JSON 解析失败重试后禁止记录原始 LLM 输出，只记录 content_len + parse_error → `knowledge/e2e-integration-patterns.md`
* **E2E并发LLM限流** : 并发创建会话需逐个检查 null + sleep 1 + 单独重试应对 LLM API 限流 → `knowledge/e2e-integration-patterns.md`
* **静态文件挂载验证盲区** : API测试通过+前端build成功≠后端托管前端文件；必须在spec/tasks中覆盖"胶水代码"横切关注点，E2E至少验证 `GET /` 返回200 → `knowledge/e2e-integration-patterns.md`
* **fetch TypeError判别** : try/catch 包裹 fetch()，`e instanceof TypeError` 区分网络故障与 HTTP 错误 → `knowledge/web-frontend-patterns.md`
* **Vue3离线检测** : `ref(navigator.onLine)` + `online`/`offline` 事件 + sticky banner → `knowledge/web-frontend-patterns.md`
