### 变更日志 — Android 客户端骨架

#### 时间线

| 时间 | 阶段 | 事件 | 备注 |
|------|------|------|------|
| 2026-05-24 | propose | cc-propose 创建 change | 阶段 2: Android 客户端 |

#### 技术决策

| 决策 | 选择 | 放弃的方案 | 原因 |
|------|------|-----------|------|
| 架构 | MVVM + Repository | MVC | Google 推荐，Compose 原生 |
| 网络 | Retrofit + OkHttp + Moshi | Ktor | SSE 支持成熟 |
| Token 存储 | EncryptedSharedPreferences | DataStore | 安全要求 |
| Markdown | 纯文本 + 简单格式化 | Markwon | 先最小可用 |
