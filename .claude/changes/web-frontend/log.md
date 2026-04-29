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

## 2026-04-30 — cc-fix

### 修复记录

| Finding | 严重度 | 修复内容 | 修改文件 |
|---------|--------|----------|----------|
| F1 | Minor | 轮数显示改为"第 N / M 轮" | ChatView.vue |
| F2 | Minor | 按钮文本改为"结束对话"/"结束中..." | ChatView.vue |
| F3 | Important | 自动结束前显示 1.5s 琥珀色提示"对话轮数已达上限，正在结束对话..." | ChatView.vue |
| F4 | Important | 历史报告列表添加点击事件和 hover 样式 | ReportView.vue |
| F5 | Important | 维度推荐失败显示琥珀色警告 banner | SetupView.vue |
| F6 | Critical | MarkdownRenderer 引入 DOMPurify 对 v-html 做消毒 | MarkdownRenderer.vue, package.json |
| F7 | Important | API 客户端先检查 res.ok 再解析 JSON，错误 JSON 解析包 try/catch | client.ts |
| F8 | Important | 区分 404（无报告）和其他错误（显示错误信息） | ReportView.vue |
| F9 | Minor | CORS Allow-Methods/Allow-Headers 移入允许源 if 分支 | server.go |
| F10 | Minor | 添加 Access-Control-Max-Age: 86400 | server.go |
| F11 | Minor | 提取 ROLE_TYPE_CUSTOM 和 DIMENSION_MODE_DERIVE 常量 | SetupView.vue |
| F12 | Important | 维度推荐 watcher 添加 300ms 防抖 | SetupView.vue |
| F13 | Minor | fallback 错误消息改为"请求失败，请稍后重试" | client.ts |

### 验证证据

| 验证项 | 结果 |
|--------|------|
| vue-tsc -b | 通过（无类型错误） |
| vite build | 通过（8 chunks 生成，1.53s） |
| go build ./... | 通过 |

### 修复决策

| 决策 | 选择 | 原因 |
|------|------|------|
| 自动结束提示方式 | inline amber message + setTimeout(1500) | 比 toast 简单，无需引入额外组件，1.5s 足够用户看到最后一条回复 |
| 维度推荐防抖 | setTimeout 300ms | 比 AbortController 简单，MVP 阶段够用 |
| XSS 防护 | DOMPurify | marked v15 已移除内置 sanitize，DOMPurify 是业界标准方案 |
| CORS Max-Age | 86400（24h） | 开发环境预检缓存，减少 OPTIONS 请求 |

## 2026-04-30 — cc-archive

### 归档记录

- **归档时间**: 2026-04-30
- **spec.status**: review → done
- **review 结论**: PASSED（13 findings 全部 fixed）
- **验证证据**: vue-tsc 通过, vite build 通过（1.50s, 8 chunks）, go build 通过, go test 95 passed

### 知识沉淀

| 知识项 | 分类 | 写入位置 |
|--------|------|----------|
| Tailwind CSS v4 Vite 集成模式 | 技术约定 | knowledge/web-frontend-patterns.md |
| marked v15 代码高亮模式 | 踩坑记录 | knowledge/web-frontend-patterns.md |
| v-html XSS 防护（DOMPurify） | 安全红线 | knowledge/web-frontend-patterns.md |
| CORS 中间件最佳实践 | 技术约定 | knowledge/web-frontend-patterns.md |
| highlight.js 分包策略 | 技术约定 | knowledge/web-frontend-patterns.md |

### 归档结论

web-frontend change 归档完成。三页面 SPA（设定→对话→报告）闭环实现，13 个 review findings 全部修复，知识沉淀 5 条。
