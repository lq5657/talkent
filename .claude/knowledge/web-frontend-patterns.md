# Web 前端模式

> 来源: `web-frontend` change 归档

## Tailwind CSS v4 Vite 集成

- Tailwind v4 不再需要 `tailwind.config.js` 和 `content` 配置
- 使用 `@tailwindcss/vite` 插件替代 PostCSS 配置：`plugins: [vue(), tailwindcss()]`
- CSS 入口只需 `@import "tailwindcss"`，Vite 插件自动检测模板文件
- 旧版 v3 的 `@tailwind base/components/utilities` 和 `tailwind.config.js` 在 v4 中无效

## marked v15 代码高亮

- marked v15 移除了 `setOptions({ highlight })` 选项，直接使用会报错
- 正确模式：`marked.use({ renderer: { code({ text, lang }) { ... } } })` 自定义 renderer
- 在 renderer 中调用 `hljs.highlight(text, { language })` 实现代码高亮
- 无语言标识时使用 `hljs.highlightAuto(text)` 兜底

## v-html XSS 防护

- marked v15 已移除内置 sanitize 选项，`marked.parse()` 输出不可直接用于 `v-html`
- 必须使用 DOMPurify：`DOMPurify.sanitize(marked.parse(content))`
- 这是 Vue 3 前端的强制安全红线，任何 Markdown 渲染场景都不能跳过

## CORS 中间件最佳实践

- Allow-Methods、Allow-Headers 必须在允许源 `if` 分支内设置，无条件设置会向非允许源泄露信息
- 添加 `Access-Control-Max-Age: 86400` 减少浏览器 OPTIONS 预检请求
- 开发环境只需允许 `localhost:5173` 和 `127.0.0.1:5173`，生产环境由反向代理处理
- Vite 开发代理（`server.proxy`）可完全避免开发期 CORS 问题

## highlight.js 分包

- highlight.js 全量包约 970KB，必须通过 Vite `manualChunks` 拆分为独立 chunk
- 配置：`build: { rollupOptions: { output: { manualChunks: { 'highlight.js': ['highlight.js'] } } } }`
- 拆分后仅在使用报告页时按需加载，不影响首屏
