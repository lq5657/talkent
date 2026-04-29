---
change_id: web-frontend
reviewed_at: 2026-04-30
reviewed_by: cc-review
stage1_status: pass
stage2_status: pass
final_status: pass
---

# Review — web-frontend

## Stage 1: Spec Compliance

### 需求覆盖矩阵

| ID | 需求 | 状态 | 备注 |
|----|------|------|------|
| S1 | 角色描述输入 | covered | |
| S2 | 场景描述输入 | covered | |
| S3 | 推荐目标列表可勾选 | covered | |
| S4 | 添加自定义目标 | covered | |
| S5 | 推荐维度 | covered | |
| S6 | 轮数限制输入 | covered | |
| S7 | 创建会话并跳转 | covered | |
| C1 | 消息流右/左对齐 | covered | |
| C2 | Enter发送, Shift+Enter换行 | covered | |
| C3 | AI回复显示 | covered | |
| C4 | 轮数计数显示 | partial | F1: 格式缺少"第...轮" |
| C5 | 手动结束对话 | partial | F2: 按钮文本为"结束"非"结束对话" |
| C6 | 轮数限制自动结束 | partial | F3: 缺少用户提示 |
| C7 | 结束后跳转报告 | covered | |
| C8 | loading和错误状态 | covered | |
| R1 | 加载最新报告 | covered | |
| R2 | Markdown渲染 | covered | |
| R3 | 维度分析卡片 | covered | |
| R4 | 手动触发分析 | covered | |
| R5 | 分析等待状态 | covered | |
| R6 | 历史报告可切换 | partial | F4: 列表不可点击切换 |
| G1 | 响应式布局 | covered | |
| G2 | 页面路由 | covered | |
| G3 | API错误统一处理 | partial | F5: 维度推荐错误被静默吞掉 |
| G4 | 加载状态 | covered | |
| CORS | 中间件 | covered | |

Stage 1 结论：**PASSED**（所有核心流程可实现，部分细节需修复）

## Stage 2: Code Quality

### 审查维度

| 维度 | 评价 |
|------|------|
| Correctness | 有若干正确性问题：API JSON解析顺序、维度推荐竞态、报告加载错误区分 |
| Readability | 良好，组件职责清晰，有少量魔法字符串 |
| Architecture | 良好，组件分解合理，API客户端封装简洁 |
| Security | **Critical**: Markdown渲染未做XSS防护，spec已识别但未实现 |
| Performance | 良好，highlight.js已分包，缺少CORS预检缓存头 |

Stage 2 结论：**PASSED with findings**（1个Critical需修复）

## Findings

### F1 | Minor | Spec — C4 轮数显示格式

**问题**: ChatView 显示 `3 / 10`，spec 验收标准要求"第 N / M 轮"

**文件**: `web/src/views/ChatView.vue:89`

**修复**: 改为 `第 {{ roundCurrent }} / {{ roundLimit }} 轮`

---

### F2 | Minor | Spec — C5 结束按钮文本

**问题**: 按钮文本为"结束"，spec 指定"结束对话"

**文件**: `web/src/views/ChatView.vue:96`

**修复**: 文本改为"结束对话"，loading 状态改为"结束中..."

---

### F3 | Important | Correctness — C6 自动结束缺少用户提示

**问题**: `is_last` 为 true 时直接调用 `endSession()` 并跳转报告页，用户没有看到"轮数已达上限"的提示，体验突兀

**文件**: `web/src/views/ChatView.vue:50-52`

**修复**: 在自动结束前显示提示（如 toast 或 inline message），或给 1-2 秒延迟让用户看到最后一条回复

---

### F4 | Important | Correctness — R6 历史报告列表不可切换

**问题**: 历史报告列表仅为展示，无点击事件，无法切换查看不同报告

**文件**: `web/src/views/ReportView.vue:114-129`

**修复**: 为每个历史报告添加点击处理，加载该报告内容显示

---

### F5 | Important | Correctness — G3 维度推荐错误被静默吞掉

**问题**: `watch(selectedGoals)` 的 catch 块完全静默，用户不知道维度推荐失败

**文件**: `web/src/views/SetupView.vue:61-63`

**修复**: 显示非阻塞性警告（如黄色提示"维度推荐失败，可继续对话"）

---

### F6 | Critical | Security — Markdown 渲染 XSS

**问题**: `MarkdownRenderer.vue` 使用 `v-html` 渲染 `marked.parse()` 输出，无任何 HTML 消毒。spec §8 已识别此风险（"marked 配置 sanitize"），但 marked v15 已移除内置 sanitize，需使用 DOMPurify

**文件**: `web/src/components/MarkdownRenderer.vue:28`

**修复**: 引入 `dompurify`，对 `marked.parse()` 输出做 `DOMPurify.sanitize()` 后再渲染

---

### F7 | Important | Correctness — API 客户端 JSON 解析在状态检查之前

**问题**: `client.ts` 先调 `res.json()` 再检查 `res.ok`。若服务器返回非 JSON 响应（如 502 HTML），会抛出不友好的 SyntaxError

**文件**: `web/src/api/client.ts:27-29`

**修复**: 先检查 `res.ok`，错误响应的 JSON 解析包在 try/catch 中，失败时回退到 `res.statusText`

---

### F8 | Important | Correctness — ReportView 将所有 API 错误视为"无报告"

**问题**: `loadReport()` 的 catch 将所有异常设 `report = null`，500 错误不应显示"生成分析"按钮而应显示错误信息

**文件**: `web/src/views/ReportView.vue:17-28`

**修复**: 区分 404（无报告）和其他错误（显示错误信息）

---

### F9 | Minor | Security — CORS 头信息泄露给非允许源

**问题**: `Access-Control-Allow-Methods` 和 `Access-Control-Allow-Headers` 无条件设置，即使源不在允许列表中

**文件**: `internal/server/server.go:38-39`

**修复**: 将 Methods 和 Headers 设置移入 `if origin == "..."` 分支内

---

### F10 | Minor | Performance — 缺少 CORS 预检缓存头

**问题**: 无 `Access-Control-Max-Age` 头，浏览器每次跨域请求前都发送 OPTIONS 预检

**文件**: `internal/server/server.go`

**修复**: 在允许源分支内添加 `Access-Control-Max-Age: 86400`

---

### F11 | Minor | Readability — API 参数魔法字符串

**问题**: `role_type: 'custom'` 和 `mode: 'derive'` 硬编码内联

**文件**: `web/src/views/SetupView.vue:55-58`

**修复**: 提取为常量

---

### F12 | Important | Correctness — 维度推荐 watcher 无防抖/取消

**问题**: `watch(selectedGoals)` 每次变更立即发请求，快速勾选多个目标时会竞态

**文件**: `web/src/views/SetupView.vue:46-66`

**修复**: 添加防抖或 AbortController 取消前序请求

---

### F13 | Minor | Spec — API 客户端 fallback 错误消息暴露状态码

**问题**: `data.error ?? \`request failed: ${res.status}\`` 暴露 HTTP 状态码

**文件**: `web/src/api/client.ts:31`

**修复**: fallback 消息改为"请求失败，请稍后重试"

---

## Findings Summary

| ID | Severity | Dimension | Status |
|----|----------|-----------|--------|
| F1 | Minor | Spec | fixed |
| F2 | Minor | Spec | fixed |
| F3 | Important | Correctness | fixed |
| F4 | Important | Correctness | fixed |
| F5 | Important | Correctness | fixed |
| F6 | Critical | Security | fixed |
| F7 | Important | Correctness | fixed |
| F8 | Important | Correctness | fixed |
| F9 | Minor | Security | fixed |
| F10 | Minor | Performance | fixed |
| F11 | Minor | Readability | fixed |
| F12 | Important | Correctness | fixed |
| F13 | Minor | Spec | fixed |

**Critical: 1 | Important: 6 | Minor: 6**

## Conclusion

**review 状态: PASS** — 所有 13 个 Findings 已修复。Critical: 0 | Important: 0 | Minor: 0

## Next Action

`cc-archive web-frontend` — 归档
