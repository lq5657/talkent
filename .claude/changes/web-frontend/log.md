# Log — web-frontend

## 2026-04-29 — cc-propose

### 决策记录

| 主题 | 决策 | 理由 |
|------|------|------|
| 前端框架 | Vue 3 + Composition API | 用户确认，组件化开发体验好，国内生态丰富 |
| 部署方式 | 前后端分离 | 兼容后续移动端 Hybrid App，Go embed 不适用于移动端 |
| 样式方案 | Tailwind CSS | 原子化 CSS，响应式断点开箱即用，减少自定义 CSS |
| Markdown 渲染 | marked + highlight.js | 轻量，满足分析报告 Markdown 渲染需求 |
| HTTP 客户端 | fetch API 封装 | MVP 阶段零额外依赖，后续可升级为 axios |
| 移动端策略 | 响应式 Web，预留 Hybrid App 路径 | 一套代码适配多端，Capacitor/Tauri 可无缝过渡 |
| CORS 方案 | 后端中间件 + Vite 开发代理 | 开发期 Vite 代理无 CORS 问题，生产期中间件或反向代理 |

### 待确认

- 生产环境 CORS 策略：同源部署还是允许特定域名？
- 是否需要 Go 后端托管前端静态文件（简化部署但增加耦合）

## 2026-04-30 — cc-apply

### 执行记录

| Task | 状态 | 关键产出 | 验证证据 |
|------|------|----------|----------|
| Task 1: Vue 3 脚手架 | done | web/ 项目初始化（package.json, vite.config.ts, tsconfig, App.vue, main.ts, style.css） | npm install 成功, vite build 产物生成, dev server 5173 端口响应 |
| Task 2: 路由与 API 客户端 | done | router/index.ts（3 路由）, api/client.ts（9 API + TypeScript 类型） | vue-tsc 类型检查通过, vite build 成功 |
| Task 3: 设定页 | done | SetupView.vue, GoalSelector.vue, DimensionList.vue | vue-tsc 通过, build 成功 |
| Task 4: 对话页 | done | ChatView.vue, MessageBubble.vue, ChatInput.vue | vue-tsc 通过, build 成功 |
| Task 5: 报告页 | done | ReportView.vue, DimensionCard.vue, MarkdownRenderer.vue | vue-tsc 通过, build 成功（marked v15 使用 renderer 替代 setOptions.highlight） |
| Task 6: CORS 中间件 | done | internal/server/server.go 新增 corsMiddleware | go test 95 passed, curl 验证 CORS 头正确（Allow-Origin 仅限 localhost:5173） |
| Task 7: 响应式布局 | done | 各页面 Tailwind 响应式类调整（md: breakpoints, grid grid-cols-1 md:grid-cols-2） | vue-tsc + vite build 通过, go test 95 passed |

### 技术决策

| 决策 | 选择 | 原因 |
|------|------|------|
| Tailwind CSS 版本 | v4 + @tailwindcss/vite 插件 | v4 无需 tailwind.config.js 和 content 配置，Vite 插件自动检测 |
| marked 高亮 | marked.use({ renderer: { code } }) + highlight.js | marked v15 移除了 setOptions({ highlight })，需用自定义 renderer |
| highlight.js 分包 | vite manualChunks | 避免 highlight.js 全量包打入 ReportView chunk（1013KB→44KB） |
| CORS 策略 | 仅允许 localhost:5173（开发） | 生产环境 CORS 由反向代理处理，后端中间件只服务于开发期 |

### 踩坑

- marked v15 的 `setOptions({ highlight })` 不再支持，需改用 `marked.use({ renderer: { code } })`
- DimensionList.vue 初版使用了 `props.xxx` 但未赋值 `const props = defineProps()`，导致 TS 报错
