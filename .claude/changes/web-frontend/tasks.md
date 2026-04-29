---
change_id: web-frontend
---

#### Task 1: Vue 3 项目脚手架与构建配置

**目标**: 初始化 Vue 3 + Vite + Tailwind CSS 项目，配置开发代理和构建

**涉及文件**: web/package.json, web/vite.config.ts, web/index.html, web/src/main.ts, web/src/App.vue, web/tailwind.config.js, web/src/style.css, .gitignore

**验收标准**:
- `npm install` 成功
- `npm run dev` 启动 Vite dev server，访问 localhost:5173 显示空白页面
- `npm run build` 生成 `web/dist/` 产物
- Vite 代理配置：`/api` 请求转发到 `http://localhost:8080`
- Tailwind CSS 生效

**验证步骤**: npm run dev 启动成功 + npm run build 产物生成 + Vite 代理转发 /api 请求到后端（V7, V8）

**测试要求**: L2 build 验证

**回退方式**: 删除 web/ 目录，恢复 .gitignore

**完成后状态**: `done`

**Baseline / Delta**: 无既有前端代码，baseline 为空

---

#### Task 2: 路由与 API 客户端

**目标**: 建立 Vue Router 路由和 HTTP 客户端封装

**涉及文件**: web/src/router/index.ts, web/src/api/client.ts

**验收标准**:
- 路由定义：`/` → SetupView, `/chat/:id` → ChatView, `/report/:id` → ReportView
- API 客户端封装：统一的 `apiGet` / `apiPost` 方法，自动处理 JSON 和错误
- API 错误统一提取 `error` 字段并抛出
- 路由切换正常工作

**验证步骤**: dev server 路由切换正常 + API 客户端调用后端 API 可返回正确响应（V5）

**测试要求**: L2

**回退方式**: 删除 router/ 和 api/ 文件

**完成后状态**: `done`

**Baseline / Delta**: 无

---

#### Task 3: 设定页（Setup Page）

**目标**: 实现角色设定页，完成角色描述→目标推荐→维度推荐→创建会话的完整流程

**涉及文件**: web/src/views/SetupView.vue, web/src/components/GoalSelector.vue, web/src/components/DimensionList.vue

**验收标准**:
- S1: 角色描述文本框，支持多行输入
- S2: 场景描述文本框
- S3: 提交后调用 recommend-goals API，展示推荐目标列表，可勾选
- S4: "添加自定义目标"入口，输入后加入已选列表
- S5: 根据角色+目标调用 recommend-dimensions API，展示维度列表
- S6: 轮数限制输入框，默认 10，范围 1-50
- S7: "开始对话"按钮，收集所有设定 → 调用 createSession → 跳转到 `/chat/{id}`
- 所有 API 调用有 loading 状态，错误有友好提示

**验证步骤**: dev server 设定页完整流程可走通（角色描述→目标推荐→维度推荐→创建会话→跳转）（V1）

**测试要求**: L2

**回退方式**: 删除 SetupView.vue、GoalSelector.vue、DimensionList.vue

**完成后状态**: `done`

**Baseline / Delta**: 无

---

#### Task 4: 对话页（Chat Page）

**目标**: 实现对话交互页，支持消息收发、轮数显示、结束对话

**涉及文件**: web/src/views/ChatView.vue, web/src/components/MessageBubble.vue, web/src/components/ChatInput.vue

**验收标准**:
- C1: 消息流展示，用户右对齐 AI 左对齐
- C2: 输入框 + 发送按钮，Enter 发送，Shift+Enter 换行
- C3: 发送后展示 AI 回复
- C4: 显示 "第 N / M 轮"
- C5: "结束对话"按钮，调用 endSession API
- C6: is_last 为 true 时自动调用 endSession
- C7: 结束成功后跳转 `/report/{id}`
- C8: loading 状态 + 错误提示

**验证步骤**: dev server 对话页完整流程可走通（发送消息→收到回复→轮数显示→结束→跳转报告）（V2）

**测试要求**: L2

**回退方式**: 删除 ChatView.vue、MessageBubble.vue、ChatInput.vue

**完成后状态**: `done`

**Baseline / Delta**: 无

---

#### Task 5: 报告页（Report Page）

**目标**: 实现分析报告展示页，支持 Markdown 渲染、维度卡片、手动触发分析

**涉及文件**: web/src/views/ReportView.vue, web/src/components/DimensionCard.vue, web/src/components/MarkdownRenderer.vue

**验收标准**:
- R1: 页面加载时调用 getReport API 获取最新报告
- R2: Markdown 内容渲染展示（marked + 代码高亮）
- R3: 维度分析卡片：名称、评分、评语、改进建议列表
- R4: 无报告时显示"生成分析"按钮，调用 analyze API
- R5: 分析生成中显示 loading，完成后刷新
- R6: 底部展示历史报告列表，可切换查看

**验证步骤**: dev server 报告页完整流程可走通（报告展示→Markdown渲染→维度卡片→手动触发分析）（V3）

**测试要求**: L2

**回退方式**: 删除 ReportView.vue、DimensionCard.vue、MarkdownRenderer.vue

**完成后状态**: `todo`

**Baseline / Delta**: 无

---

#### Task 6: CORS 中间件与后端适配

**目标**: 后端新增 CORS 中间件，支持前后端分离开发

**涉及文件**: internal/server/server.go

**验收标准**:
- 新增 CORS 中间件，允许 `http://localhost:5173`（开发）和同源（生产）的跨域请求
- 允许的 Methods：GET, POST, OPTIONS
- 允许的 Headers：Content-Type
- OPTIONS 预检请求正确响应
- 不影响现有 API 功能（现有测试仍通过）

**验证步骤**: go test ./internal/server/... 通过 + curl 验证 CORS 响应头（Access-Control-Allow-Origin 等）（V6）

**测试要求**: L2

**回退方式**: 回退 server.go 变更

**完成后状态**: `done`

**Baseline / Delta**: 修改既有 server.go，需 baseline 比对

---

#### Task 7: 响应式布局与整体集成

**目标**: 适配移动端响应式布局，端到端流程验证

**涉及文件**: 各 View/Component 的 Tailwind 响应式类调整

**验收标准**:
- G1: 桌面（≥1024px）三栏/宽布局，手机（<768px）单栏/窄布局
- 完整流程验证：设定 → 对话 → 结束 → 报告，端到端可走通
- 所有页面在不同宽度下布局不溢出、不截断

**验证步骤**: 浏览器多宽度（1024px / 768px / 375px）手工验证每个页面和完整流程（V4）

**测试要求**: L4 manual

**回退方式**: 回退 Tailwind 响应式类调整

**完成后状态**: `done`

**Baseline / Delta**: Task 3-5 完成后的布局快照
