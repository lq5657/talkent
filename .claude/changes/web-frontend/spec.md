---
change_id: web-frontend
status: review
depends_on:
  - scaffold-project
  - llm-client
  - role-and-goal
  - chat-session
  - analysis-engine
parallel_safe: false
branch: feat/web-frontend
created: 2026-04-29
updated: 2026-04-29
complexity: large
min_validation_level: L3
---

# Web 前端：三页面 SPA 应用

## 1. 目标

为 Talkent 提供 Web 前端，让用户在浏览器中完成"角色设定 → 对话训练 → 查看分析报告"的完整闭环。

## 2. 用户与场景

- 目标用户：任何想通过角色扮演对话提升表达能力的人
- 高频动作：设定角色 → 选择训练目标 → 开始对话 → 结束对话 → 查看分析报告
- 访问设备：桌面浏览器（MVP 首选）、手机浏览器（响应式适配）

## 3. 功能需求

### 3.1 设定页（Setup Page）

| ID | 需求 | 验收标准 | 对应 API |
|----|------|----------|----------|
| S1 | 用户输入角色描述 | 文本框可输入，提交后调用目标推荐 | `POST /api/roles/recommend-goals` |
| S2 | 用户输入场景描述 | 文本框可输入，随角色描述一起提交 | 同上（createSession 的 scenario 字段） |
| S3 | 展示推荐训练目标列表 | 调用 API 后展示目标列表，每项可勾选 | 同上 |
| S4 | 用户可补充自定义目标 | 提供"添加目标"入口，输入后加入已选列表 | 无 API，前端本地状态 |
| S5 | 展示推荐分析维度 | 根据角色+目标调用维度推荐 API | `POST /api/roles/recommend-dimensions` |
| S6 | 设定对话轮数限制 | 数字输入框，默认值 10，范围 1-50 | 无 API，传给 createSession |
| S7 | 确认设定并创建会话 | 收集角色、场景、目标、维度、轮数 → 调用创建会话 API → 跳转对话页 | `POST /api/sessions` |

### 3.2 对话页（Chat Page）

| ID | 需求 | 验收标准 | 对应 API |
|----|------|----------|----------|
| C1 | 显示对话消息流 | 用户消息右对齐、AI 消息左对齐，按时间顺序排列 | 前端本地状态 |
| C2 | 用户输入消息并发送 | 输入框 + 发送按钮，Enter 快捷发送，Shift+Enter 换行 | `POST /api/sessions/{id}/chat` |
| C3 | 显示 AI 回复 | 发送后展示 AI 回复内容 | 同上 |
| C4 | 显示当前轮数和限制 | 底部或顶部展示 "第 N / M 轮" | chatResponse.round_info |
| C5 | 手动结束对话 | "结束对话"按钮，点击后调用结束 API | `POST /api/sessions/{id}/end` |
| C6 | 达到轮数限制自动结束 | 轮数用尽后自动调用结束 API 并提示 | chatResponse.round_info.is_last |
| C7 | 结束后跳转报告页 | 结束成功后自动跳转到分析报告页 | 前端路由 |
| C8 | 加载状态和错误提示 | AI 回复期间显示 loading，API 失败显示错误提示 | 前端 UI |

### 3.3 报告页（Report Page）

| ID | 需求 | 验收标准 | 对应 API |
|----|------|----------|----------|
| R1 | 展示分析报告 | 页面加载时获取最新报告 | `GET /api/sessions/{id}/report` |
| R2 | 渲染 Markdown 报告 | 使用 Markdown 渲染器展示报告正文 | 同上 markdown 字段 |
| R3 | 展示维度分析卡片 | 每个维度显示：名称、评分、评语、改进建议 | 同上 dimensions 字段 |
| R4 | 手动触发分析 | 如果对话结束但无报告，提供"生成分析"按钮 | `POST /api/sessions/{id}/analyze` |
| R5 | 分析生成中的等待状态 | 点击生成后显示 loading，完成后刷新展示 | 前端轮询或等待 |
| R6 | 历史报告列表 | 展示该会话的历史分析报告列表 | `GET /api/sessions/{id}/reports` |

### 3.4 通用需求

| ID | 需求 | 验收标准 |
|----|------|----------|
| G1 | 响应式布局 | 桌面（≥1024px）和手机（<768px）均可正常使用 |
| G2 | 页面路由 | 设定页 / → 对话页 /chat/:id → 报告页 /report/:id |
| G3 | API 错误统一处理 | 所有 API 错误展示友好提示，不暴露技术细节 |
| G4 | 加载状态 | 所有异步操作有 loading 指示 |

## 4. 技术方案

### 4.1 技术栈

| 项目 | 选型 | 理由 |
|------|------|------|
| 框架 | Vue 3 + Composition API | 用户确认，组件化开发体验好 |
| 构建工具 | Vite | Vue 3 生态标配，HMR 快 |
| 样式 | Tailwind CSS | 原子化 CSS，响应式断点开箱即用 |
| 路由 | Vue Router 4 | Vue 官方路由 |
| HTTP 客户端 | fetch API（封装） | MVP 阶段足够，无额外依赖 |
| Markdown 渲染 | marked + highlight.js | 轻量 Markdown 渲染 + 代码高亮 |

### 4.2 部署架构

- 开发期：Vite dev server（5173）+ `vite.config.ts` 代理 `/api` 到 Go 后端（8080）
- 生产期：`npm run build` 产物部署到独立静态服务器，API 通过反向代理或 CORS 访问
- 前后端分离部署，预留 Hybrid App（Capacitor/Tauri）过渡路径

### 4.3 兼容性

- 本次变更不修改后端 API 接口
- 后端需新增 CORS 中间件以支持前后端分离开发
- 后端需新增静态文件回退路由（可选，用于生产部署简化）

## 5. 影响范围

### 5.1 新增文件

| 路径 | 用途 |
|------|------|
| `web/` | Vue 3 项目根目录 |
| `web/package.json` | 前端依赖和脚本 |
| `web/vite.config.ts` | Vite 构建和代理配置 |
| `web/index.html` | SPA 入口 HTML |
| `web/src/main.ts` | Vue 应用入口 |
| `web/src/App.vue` | 根组件 + 路由视图 |
| `web/src/router/index.ts` | 路由定义 |
| `web/src/api/client.ts` | HTTP 客户端封装 |
| `web/src/views/SetupView.vue` | 设定页 |
| `web/src/views/ChatView.vue` | 对话页 |
| `web/src/views/ReportView.vue` | 报告页 |
| `web/src/components/*.vue` | 可复用组件 |
| `web/tailwind.config.js` | Tailwind 配置 |

### 5.2 修改文件

| 路径 | 修改内容 | 兼容性 |
|------|----------|--------|
| `internal/server/server.go` | 新增 CORS 中间件 | compatible_addition |
| `cmd/server/main.go` | 可选：新增静态文件回退路由 | compatible_addition |
| `.gitignore` | 新增 `web/node_modules/`、`web/dist/` | 无影响 |

## 6. 不做（Out of Scope）

- 用户认证/登录
- 对话历史列表页
- 流式对话输出（SSE）
- 语音输入/输出
- PWA 离线支持
- 国际化（i18n）
- 暗色模式
- E2E 自动化测试（由 `e2e-integration` change 覆盖）

## 7. 验证映射

| 编号 | 需求项 / 风险点 | 最低验证等级 | 证据类型 | 建议验证动作 | 对应 Task | 闭环状态 |
|------|------------------|--------------|----------|--------------|-----------|----------|
| V1 | S1-S7 设定页完整流程 | L2 | package | dev server 手工验证设定页全流程 | Task 3 | todo |
| V2 | C1-C8 对话页完整流程 | L2 | package | dev server 手工验证对话页全流程 | Task 4 | todo |
| V3 | R1-R6 报告页完整流程 | L2 | package | dev server 手工验证报告页全流程 | Task 5 | todo |
| V4 | G1 响应式布局 | L4 | manual | 浏览器多宽度手工验证 | Task 7 | todo |
| V5 | G2-G4 路由和通用交互 | L2 | package | dev server 路由切换验证 | Task 2 | todo |
| V6 | CORS 中间件正确性 | L2 | package | go test + CORS 响应头验证 | Task 6 | todo |
| V7 | API 代理开发期可用 | L3 | chain | Vite 代理 + Go 后端联调验证 | Task 1 | todo |
| V8 | 生产构建产物可用 | L2 | package | npm run build 产物验证 | Task 1 | todo |

## 8. 风险

| 风险 | 等级 | 缓解 |
|------|------|------|
| LLM API 调用慢导致对话体验差 | 中 | 前端 loading 状态 + 合理超时提示 |
| Markdown 渲染 XSS | 低 | marked 配置 sanitize |
| CORS 配置不当导致前后端无法通信 | 低 | 开发期 Vite 代理兜底，生产期反向代理 |

## 9. 拆分理由

本次 change 虽然涉及 3 个页面，但共享同一技术栈、同一部署架构、同一 API 层，且只有一个验收路径（设定→对话→报告闭环）。拆成多个 change 会增加跨 change 状态同步成本，不拆更利于迭代和验证。规模判定为 L（跨域、多页面、新依赖栈），但验收路径统一，接受为单 change。

## 10. 确认记录（HARD-GATE）

* **confirmed_at**：2026-04-29
* **confirmed_by**：lq5657
* **confirmed_spec_revision**：`spec.md` @ 2026-04-29，3 页面（设定/对话/报告），Vue 3 + Vite + Tailwind，前后端分离部署，V1-V8 验证映射（7 列），S1-S7/C1-C8/R1-R6/G1-G4 功能点 + CORS 中间件 + 响应式布局
* **confirmed_tasks_revision**：`tasks.md` @ 2026-04-29，7 个 Task（脚手架→路由/API→设定页→对话页→报告页→CORS→响应式集成）
* **confirmed_scope**：`web/` Vue 3 项目（package.json, vite.config.ts, views, components, router, api）；`internal/server/server.go` CORS 中间件；`.gitignore` 前端忽略规则；不含用户认证/历史列表/SSE/语音/PWA/i18n/暗色模式
* **accepted_risks**：LLM API 慢影响对话体验（loading + 超时提示）；Markdown XSS（marked sanitize）；CORS 配置不当（Vite 代理兜底 + 反向代理）
* **human_review_required**：false
* **human_review_status**：not_required
