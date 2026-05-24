---
change_id: android-core-shell
status: review
depends_on: []
parallel_safe: true
branch: feat/android-core-shell
created: 2026-05-24
updated: 2026-05-24
complexity: complex
proposal_profile: staged
verification_level: L3
evidence_types: [package, chain, manual]
---

### Android 客户端 — 项目骨架 + 对话 + 报告

#### 0.1 需求收敛记录

`原始诉求` → 支持安卓客户端 app 上使用 Talkent（全部功能）
→ `阶段 1: android-backend-ready` → 后端 JWT 认证 + Web 登录 ✅ 已归档
→ `阶段 2: android-core-shell` → Android 项目骨架 + 3 页面文字对话 + 报告 + 设置

#### 0.2 Proposal Profile

| Profile | 使用条件 | 本次 |
|---------|----------|------|
| staged | 需求→方案→scope 三段确认 | ✓ |

#### 1. 背景与目标

`android-backend-ready` 已完成后端 JWT 认证和 Web 登录适配。本 change 创建 Android 原生客户端，覆盖角色设定、文字对话（SSE 流式）和分析报告三个核心场景，使用 Kotlin + Jetpack Compose + MVVM 架构。

#### 1.0 路线图对齐

```
✅ android-backend-ready (done)
⬜ android-core-shell    ← 本次
⬜ android-voice         ← STT/TTS
⬜ android-offline-cache ← Room
```

#### 1.1 本次不做

- 语音交互 (STT/TTS) → `android-voice`
- 离线缓存 (Room) → `android-offline-cache`
- 用户注册/多用户
- 推送通知
- Markdown 完整渲染（仅展示纯文本/简单格式）

#### 2. 代码现状

后端 API 已就绪且已验证：

| 端点 | 认证 | 用途 |
|------|------|------|
| POST /api/auth/login | 无 | 获取 JWT token |
| POST /api/auth/refresh | 无 | 刷新 access token |
| POST /api/roles/recommend-goals | Bearer | 推荐训练目标 |
| POST /api/roles/recommend-dimensions | Bearer | 推荐分析维度 |
| POST /api/sessions | Bearer | 创建会话 |
| POST /api/sessions/{id}/chat | Bearer | 发送消息 |
| GET /api/sessions/{id}/chat/stream | query token | SSE 流式回复 |
| POST /api/sessions/{id}/end | Bearer | 结束会话 |
| GET /api/sessions/{id} | Bearer | 获取会话详情 |
| POST /api/sessions/{id}/analyze | Bearer | 触发分析 |
| GET /api/sessions/{id}/report | Bearer | 获取报告 |

Android 客户端目录 `android/` 当前不存在，从零创建。

#### 3. 功能点

* [ ] F1: 登录 — username/password 输入 → POST /api/auth/login → 存储 token → 跳转 Setup
* [ ] F2: 角色设定 — 输入描述 + 场景 → 推荐目标 → 选择目标 → 推荐维度 → 创建会话
* [ ] F3: 对话训练 — 文字输入 → SSE 流式回复 → 轮数显示 → 结束会话 → 触发分析
* [ ] F4: 分析报告 — 维度卡片 + 报告文本展示
* [ ] F5: 设置 — 后端地址配置 + 登录状态/登出
* [ ] F6: 自动 token refresh — 401 → refresh → 重试请求 → 失败则跳转登录

#### 4. 业务规则

- Token 存储在 EncryptedSharedPreferences，应用重启不丢失
- 后端地址存储在 SharedPreferences，默认 `http://10.0.2.2:8080`（模拟器本机映射）
- 对话轮数达到上限后自动结束，提示用户触发分析
- 聊天流式接收 token 逐字显示，SSE 完成时替换为完整回复
- 报告维度分数以卡片形式展示，Markdown 正文渲染为格式化文本

#### 5. 数据变更

* **是否涉及 migration**：否（新项目，无旧数据）

#### 6. 接口变更

* **是否涉及对外契约变更**：否（纯消费现有 API）

#### 7. 影响范围

| 目录 | 影响 | 类型 |
|------|------|------|
| `android/` | 新建完整 Android 项目 | 新增 |

#### 7.1 配置变更

* **是否涉及配置项或环境变量变更**：否（Android 端配置通过 App 内设置页面管理）

#### 8. 风险与关注点

| 类型 | 描述 | 处理方式 |
|------|------|----------|
| 兼容 | 模拟器 vs 真机网络地址差异 | 默认 10.0.2.2（模拟器），真机通过设置页面配置局域网 IP |
| SSE | OkHttp SSE 流式接收实现 | 使用 OkHttp + `response.body.source()` 逐行读取 SSE 流 |
| Markdown | Android 无自带 Markdown 渲染 | 初期展示纯文本，后续可集成 Markwon 等库 |
| 构建 | 首次 Gradle 同步可能较慢 | 提前配置国内镜像（如有需要） |

#### 9. 测试策略

* **测试范围**：ViewModel 单元测试 + API Repository 测试 + 手工 UI 验证
* **最低验证等级**：L3（ViewModel + Repository chain）
* **验证证据要求**：
  - L2: ViewModel 和 Repository 的 JVM 单元测试
  - L3: 连接真实后端端到端手动测试 (Setup → Chat → Report)
  - L4: APK 安装到真机/模拟器运行验证

#### 9.1 需求-验证映射

| 编号 | 需求项 | 最低验证等级 | 证据类型 | 建议验证动作 | 对应 Task | 闭环状态 |
|------|--------|-------------|----------|-------------|-----------|----------|
| V1 | 登录 + token 管理 | L3 | chain | login → token → 跳转 Setup | T3, T5 | todo |
| V2 | 角色设定流程 | L3 | chain | 输入角色描述 → 推荐目标 → 创建会话 | T4 | todo |
| V3 | SSE 流式对话 | L3 | chain | 发送消息 → 流式接收 → 逐字显示 | T5 | todo |
| V4 | 分析报告展示 | L3 | chain | 结束会话 → 分析 → 查看报告 | T6 | todo |
| V5 | 设置页面 + 后端地址 | L4 | manual | 修改地址 → API 调用成功 | T8 | todo |
| V6 | Token 自动 refresh | L2 | package | 401 → refresh → 重试 | T3 | todo |

#### 10. 待澄清

* 无——所有关键决策已在交互中确认

#### 10.1 风险决策

| 决策风险 | 可选处理路径 | 推荐路径 | 用户选择 |
|----------|--------------|----------|----------|
| 项目位置 | android/ 目录 / 独立仓库 | android/ | android/ |
| 功能范围 | 全部 3 页 / 部分 | 全部 3 页 | 全部 3 页 |
| 后端地址 | App 内配置 / 硬编码 / 向导 | App 内配置 | App 内配置 |

#### 12. 技术决策

| 决策 | 选择 | 放弃的方案 | 原因 |
|------|------|-----------|------|
| UI 框架 | Jetpack Compose + Material 3 | XML Views | 声明式，与现有 Vue 3 理念一致 |
| 架构 | MVVM + Repository | MVC | Google 推荐，Compose 原生支持 |
| 网络 | Retrofit + OkHttp + Moshi | Ktor | 生态成熟，SSE 支持好 |
| Token 存储 | EncryptedSharedPreferences | DataStore | 认证 token 需要加密存储 |
| Markdown | 纯文本 + 简单格式化 | Markwon/Compose Rich Text | 先最小可用，后续迭代 |
| 最低 SDK | API 26 (Android 8.0) | — | 前期已确认 |

#### 15. 确认记录（HARD-GATE）

* **confirmed_at**：2026-05-24
* **confirmed_by**：user
* **confirmed_spec_revision**：2026-05-24
* **confirmed_tasks_revision**：2026-05-24
* **confirmed_scope**：Android 项目骨架 + 4 页面 (Setup/Chat/Report/Settings)，10 task，6 wave
* **resolved_risk_decisions**：android/ monorepo，全部 3 页面，App 内配置后端地址
* **accepted_residual_risks**：Markdown 纯文本展示，首次 Gradle 同步慢
* **human_review_required**：true
* **human_review_status**：approved
