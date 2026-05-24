---
change_id: android-backend-ready
status: review
depends_on: []
parallel_safe: true
branch: feat/android-backend-ready
created: 2026-05-24
updated: 2026-05-24 (all tasks done)
complexity: medium
proposal_profile: staged
verification_level: L3
evidence_types: [package, chain]
---

### 后端 Android 适配 — JWT 认证 + 网络可达

#### 0.1 需求收敛记录

`原始诉求` → `支持安卓客户端app上使用 Talkent` 
→ `关键澄清` → Android 原生 (Kotlin)、全部功能、保留语音、局域网+公网双模式部署、JWT 认证
→ `收敛后的目标` → 后端先行：添加 JWT 认证中间件，使 API 可通过网络安全访问，Web 前端同步支持登录

#### 0.2 Proposal Profile

| Profile | 使用条件 | 本次 |
|---------|----------|------|
| staged | 多阶段确认交互，需求→方案→scope 逐段确认 | ✓ |

#### 1. 背景与目标

Talkent 当前为单用户本地应用，无认证机制。为支持 Android 客户端通过网络访问，需要：
1. 添加 JWT 认证层，保护所有 /api/* 路由
2. Web 前端同步支持登录/Token 管理
3. 后端已默认监听 0.0.0.0:8080，网络可达性已满足

#### 1.0 路线图对齐

本 change 是 Android 支持计划的第一阶段：

```
android-backend-ready (本 change) → android-core-shell → android-voice → android-offline-cache
```

| 后续 change | 内容 | 预估规模 |
|-------------|------|----------|
| android-core-shell | Android 项目 (Kotlin + Compose) + 对话 + 报告 | M-L |
| android-voice | Android STT/TTS + SSE 流式 | M |
| android-offline-cache | Room 离线缓存历史/报告 | S-M |

#### 1.1 本次不做

- Android 客户端代码（后续 change）
- 用户注册/多用户（仅单用户认证）
- OAuth2/第三方登录
- 密码加密存储（第一期明文配置，后续迭代）

#### 2. 代码现状

##### 2.1 相关入口与链路

| 入口 | 路径 | 说明 |
|------|------|------|
| 配置加载 | `internal/config/config.go:51 Load()` | 已有 11 个 TALKENT_* 环境变量 |
| HTTP 中间件栈 | `internal/server/server.go:30-34` | RequestID → Timeout → Recovery → CORS → mux |
| 路由注册 | `cmd/server/main.go:71-79` | 7 个 API 路由 + 静态文件 + SPA |
| 前端 API 客户端 | `web/src/api/client.ts` | 无认证 header，无 token 管理 |
| 前端路由 | `web/src/router/index.ts` | 3 个视图 (Setup/Chat/Report)，无登录页 |

##### 2.2 现有实现

- 中间件栈从外到内: RequestID → Timeout → Recovery → CORS → mux
- 所有 /api/* 路由无认证保护
- 健康检查 `GET /health` 仅用于 DB 连通性
- Server 默认监听 `0.0.0.0:8080`（已满足网络可达）
- 无 auth 相关依赖

##### 2.3 发现与风险

- CORS 中间件在 auth 之前被调用，OPTIONS 预检需要跳过认证
- chat/stream (SSE) 需要携带 token——通过 query param 传递（EventSource 不支持自定义 header）
- 前端未使用 Pinia 等状态管理——token 管理需额外设计
- Server.Host 已默认 0.0.0.0，无需修改

#### 3. 功能点

* [ ] F1：`POST /api/auth/login` — 验证凭据，返回 JWT access_token + refresh_token
* [ ] F2：`POST /api/auth/refresh` — 用 refresh_token 换新 access_token
* [ ] F3：Auth 中间件 — 验证所有 /api/* 路由的 Bearer Token（/api/auth/* 除外）
* [ ] F4：Web 登录页 — 用户名/密码输入 → token 存储 → 路由守卫
* [ ] F5：Web API 客户端 — 自动携带 Authorization header，401 时尝试 refresh

#### 4. 业务规则

- 登录凭据通过配置文件/环境变量设置（单用户模式）
- Access token 有效期 1 小时，refresh token 有效期 7 天
- /health 和 /api/auth/* 不校验认证
- OPTIONS 预检请求不校验认证（CORS 兼容）
- Token 无效/过期返回 401 `{"error":"unauthorized"}`
- Refresh token 过期需重新登录
- 无用户注册——凭据仅从配置读取

#### 5. 数据变更

* **是否涉及 migration**：否
* **变更类型**：无（无数据库 Schema 变更，仅内存中 JWT 验证）

#### 6. 接口变更

* **是否涉及对外契约变更**：是
* **兼容性分类**：compatible_addition（新增认证要求，现有调用方需加 header）
* **客户端/消费者影响**：Web 前端需添加登录流程；Android 客户端需实现登录
* **迁移路径**：无旧客户端需要迁移（当前仅 Web 前端，同步更新）
* **回滚影响**：移除 auth 中间件即可回退到无认证状态

| 操作 | 接口 | 方法 | 变更内容 | 兼容性 |
|------|------|------|----------|--------|
| 新增 | /api/auth/login | POST | 登录接口 | compatible_addition |
| 新增 | /api/auth/refresh | POST | Token 刷新 | compatible_addition |
| 修改 | /api/* (全部) | * | 新增 Authorization header 要求 | compatible_adjustment |

#### 7. 影响范围

| 模块 | 影响 | 类型 |
|------|------|------|
| `internal/config/` | 新增 AuthConfig (4 个字段) | 修改 |
| `internal/auth/` | 新包：JWT 生成/验证 + 中间件 | 新增 |
| `internal/server/` | 中间件栈加入 Auth | 修改 |
| `cmd/server/` | 路由注册增加 auth 路由 | 修改 |
| `web/src/api/client.ts` | 添加 Authorization header + refresh 逻辑 | 修改 |
| `web/src/router/` | 添加登录页路由 + 导航守卫 | 修改 |
| `web/src/views/` | 新增 LoginView.vue | 新增 |
| `config.yaml` | 新增 auth 配置段 | 修改 |

#### 7.1 配置变更

* **是否涉及配置项或环境变量变更**：是
* **新增/变更配置项**：

| 配置项 | 环境变量 | 默认值 | 必填 | 说明 |
|--------|----------|--------|------|------|
| auth.username | TALKENT_AUTH_USERNAME | admin | 是 | 登录用户名 |
| auth.password | TALKENT_AUTH_PASSWORD | admin | 是 | 登录密码（建议环境变量注入） |
| auth.jwt_secret | TALKENT_AUTH_JWT_SECRET | change-me-in-production | 是 | JWT 签名密钥 |
| auth.jwt_expiry | TALKENT_AUTH_JWT_EXPIRY | 1h | 否 | Access token 有效期 |

* **回滚影响**：移除 auth 配置段 + 移除 Auth 中间件即可回退

#### 8. 风险与关注点

| 类型 | 描述 | 处理方式 |
|------|------|----------|
| 安全 | JWT secret 硬编码默认值 | 默认值设为 change-me-in-production，生产环境通过环境变量注入 |
| 安全 | 密码明文存储在 config.yaml | 文档提示用 TALKENT_AUTH_PASSWORD 环境变量代替 |
| 兼容 | SSE chat/stream 不支持自定义 header | 通过 query param `?token=xxx` 传递（中间件优先检查 query param） |
| 兼容 | CORS OPTIONS 预检 | Auth 中间件在 OPTIONS 请求时跳过认证 |
| 可用 | 登录页刷新后 token 丢失 | localStorage 持久化 token |

#### 8.1 日志与可观测性

* **是否新增运行时日志点**：是
* **涉及入口**：POST /api/auth/login (info: 登录成功/失败)，Auth 中间件 (warn: token 无效/过期)
* **关键字段**：request_id, username (登录), reason (认证失败原因)
* **发布后观察窗口**：首次部署后观察 401 错误率

#### 9. 测试策略

* **测试范围**：auth package 单元测试 + 中间件链测试 + E2E 登录流程
* **最低验证等级**：L3（auth service + middleware chain + Web login flow）
* **验证证据要求**：
  - L2: `go test ./internal/auth/...` 覆盖 JWT 生成/验证/过期
  - L3: chain 验证：login → access_token → request /api/sessions → 200
  - L4: E2E curl 新增 scenario-6-auth.sh 覆盖登录+认证+refresh

#### 9.1 需求-验证映射

| 编号 | 需求项 / 风险点 | 最低验证等级 | 证据类型 | 建议验证动作 | 对应 Task |
|------|------------------|--------------|----------|--------------|-----------|
| V1 | JWT 生成/验证正确性 | L2 | package | go test internal/auth/ | T1 | todo |
| V2 | Auth 中间件拦截未认证请求 | L2 | package | middleware test: 无 token→401 | T2 | todo |
| V3 | 登录→token→API 全链路 | L3 | chain | curl login → token → GET /api/sessions → 200 | T3 | todo |
| V4 | Web 登录页 + token 管理 | L4 | manual | 浏览器登录 → 跳转 → 对话 → 刷新不丢失 | T4 | todo |
| V5 | SSE stream 认证 (query token) | L3 | chain | curl with token query param → stream | T5 | todo |
| V6 | Token 过期 + refresh 流程 | L2 | package | go test: 过期 token→401, refresh→新token | T1 | todo |

#### 9.2 发布与回滚

* **发布方式**：直接发布（单服务，无灰度）
* **回滚路径**：代码回滚 / 移除 Auth 中间件代码回退
* **发布后观察窗口**：1 小时（观察 401 错误率）

#### 10. 待澄清

* 无——所有关键决策已在交互中确认

#### 10.1 风险决策

| 决策风险 | 可选处理路径 | 推荐路径 | 用户选择 |
|----------|--------------|----------|----------|
| 认证方案 | API Key / JWT / 无认证 | JWT | JWT |
| Android 最低 SDK | API 24 / 26 / 31 | API 26 | API 26 |
| 部署形态 | 局域网 / 公网 / 都要 | 都要 | 都要 |

#### 11. 方案比较

| 方案 | 是否采用 | 适用前提 | 采用 / 放弃原因 |
|------|----------|----------|-----------------|
| JWT Token | ✓ | 无状态认证，适合 API 服务 | 用户选择 |
| API Key | ✗ | 更简单，但无过期机制 | 用户未选择 |
| 无认证 | ✗ | 仅本地使用 | 网络暴露下不安全 |

#### 12. 技术决策

| 决策 | 选择 | 放弃的方案 | 原因 |
|------|------|-----------|------|
| JWT 库 | golang-jwt/jwt/v5 | 自实现 HMAC | 标准库，审计充分 |
| Token 传递（SSE） | Query param | 自定义 header | EventSource API 不支持自定义 header |
| Web 状态管理 | localStorage + 简单 composable | Pinia | 最小变更，不引入新依赖 |
| 凭据来源 | 配置文件 + 环境变量 | 数据库 | 单用户模式，无注册需求 |

#### 15. 确认记录（HARD-GATE）

* **confirmed_at**：2026-05-24
* **confirmed_by**：user
* **confirmed_spec_revision**：2026-05-24
* **confirmed_tasks_revision**：2026-05-24
* **confirmed_scope**：后端 JWT 认证 + Web 登录适配，7 个 task，5 个 wave
* **resolved_risk_decisions**：JWT Token 方案，SSE query param 传递 token，localStorage 存储，配置文件凭据
* **accepted_residual_risks**：JWT secret 默认值不安全（生产环境变量覆盖），密码明文配置（环境变量注入），单用户模式
* **human_review_required**：true
* **human_review_status**：approved
