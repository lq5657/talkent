---
change_id: android-backend-ready
created: 2026-05-24
updated: 2026-05-24
---

### 任务拆分 — 后端 Android 适配

#### 前置条件

* [ ] `spec.md` 已确认且 `status = propose`
* [ ] `depends_on` 空，无前置 change

#### 依赖 / Wave 总览

```
Wave 1 (独立并行): T1 (auth package) + T6 (config)
       │
Wave 2 (依赖 T1+T6): T2 (auth middleware)
       │
Wave 3 (依赖 T2): T3 (server integration + routes)
       │
Wave 4 (依赖 T3): T4 (Web login page) + T5 (Web API client)
       │
Wave 5 (依赖 T4+T5): T7 (E2E verification)
```

#### 变更影响概览

##### 文件变更清单

| 文件 | 操作 | 涉及 Task | 说明 |
|------|------|-----------|------|
| internal/auth/jwt.go | 新增 | T1 | JWT 生成/验证/解析 |
| internal/auth/jwt_test.go | 新增 | T1 | JWT 单元测试 |
| internal/auth/middleware.go | 新增 | T2 | Auth HTTP 中间件 |
| internal/auth/middleware_test.go | 新增 | T2 | 中间件测试 |
| internal/auth/handler.go | 新增 | T3 | login + refresh handler |
| internal/config/config.go | 修改 | T6 | 新增 AuthConfig |
| internal/server/server.go | 修改 | T3 | 中间件栈加入 Auth + login 路由 |
| cmd/server/main.go | 修改 | T3 | 传递 auth handler |
| web/src/views/LoginView.vue | 新增 | T4 | 登录页面 |
| web/src/router/index.ts | 修改 | T4 | 新增 /login 路由 + 导航守卫 |
| web/src/api/client.ts | 修改 | T5 | Authorization header + refresh |
| web/src/composables/useAuth.ts | 新增 | T5 | Token 管理 composable |
| config.yaml | 修改 | T6 | 新增 auth 配置段 |
| test/e2e/curl/scenario-6-auth.sh | 新增 | T7 | E2E 认证验证 |

##### 受影响接口 / 调用方

| 接口 / 入口 | 变更类型 | 上游调用方 | 下游依赖 | 涉及 Task |
|-------------|----------|------------|----------|-----------|
| POST /api/auth/login | 新增 | Web/Android 客户端 | config.AuthConfig | T3 |
| POST /api/auth/refresh | 新增 | Web/Android 客户端 | JWT service | T3 |
| 所有 /api/* 路由 | 修改（需认证） | Web/Android 客户端 | Auth middleware | T2, T3 |
| web/api/client.ts | 修改 | 所有 Vue 视图/组件 | Auth composable | T5 |
| web/router/index.ts | 修改 | Vue app | LoginView | T4 |

#### Task 1: auth package — JWT service

* **目标**: 实现 JWT access/refresh token 的生成、验证、解析
* **不包含范围**: HTTP 中间件、login handler
* **涉及文件**: `internal/auth/jwt.go`, `internal/auth/jwt_test.go`
* **关键签名**:
  - `func NewJWTService(secret string, accessExpiry, refreshExpiry time.Duration) *JWTService`
  - `func (s *JWTService) GenerateAccessToken(username string) (string, error)`
  - `func (s *JWTService) GenerateRefreshToken(username string) (string, error)`
  - `func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error)`
  - `func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error)`
* **验收标准**: access token 过期返回 error；refresh token 可刷新；无效 token 返回 error
* **验证步骤**: `go test ./internal/auth/ -v`
* **测试要求**: 覆盖 happy path + 过期 + 无效 token + 错误 secret
* **测试要求**: 覆盖 happy path + 过期 + 无效 token + 错误 secret
* **依赖 / Wave**: Wave 1（无依赖）
* **回退方式**: 删除 auth/ 目录
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A（新包，无基线）

#### Task 2: Auth middleware

* **目标**: 实现 HTTP 中间件，验证 /api/* 路由的 Bearer Token 或 query param token
* **不包含范围**: login/refresh handler
* **涉及文件**: `internal/auth/middleware.go`, `internal/auth/middleware_test.go`
* **关键签名**:
  - `func AuthMiddleware(jwtSvc *JWTService, logger *slog.Logger) func(http.Handler) http.Handler`
* **验收标准**:
  - 无 token → 401 `{"error":"unauthorized"}`
  - 无效 token → 401
  - 过期 token → 401（区分 invalid 和 expired 以便 client refresh）
  - 有效 token → 注入 username 到 context，放行
  - OPTIONS 请求 → 跳过认证
  - /health 和 /api/auth/* → 跳过认证
  - token 可从 query param `?token=xxx` 获取（SSE 兼容）
* **验证步骤**: `go test ./internal/auth/ -v -run TestAuthMiddleware`
* **测试要求**: httptest 覆盖所有 6 个验收标准
* **测试要求**: httptest 覆盖所有 6 个验收标准
* **依赖 / Wave**: Wave 2（依赖 T1, T6）
* **回退方式**: 从中间件栈移除
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A（新中间件，无基线）

#### Task 3: Server 集成 + Login/Refresh Handler

* **目标**: 将 auth 中间件加入 server 中间件栈，注册 login/refresh 路由
* **不包含范围**: Web 前端
* **涉及文件**: `internal/auth/handler.go`, `internal/server/server.go`, `cmd/server/main.go`
* **验收标准**:
  - `POST /api/auth/login` 接受 username/password，返回 token pair
  - `POST /api/auth/refresh` 接受 refresh_token，返回新 access_token
  - Auth 中间件位于 CORS 之后、mux 之前
  - 中间件栈顺序: RequestID → Timeout → Recovery → CORS → Auth → mux
* **验证步骤**: curl 测试 login → 200 + token；curl 无 token 访问 /api/sessions → 401
* **测试要求**: handler 测试通过 httptest
* **测试要求**: handler test 通过 httptest
* **依赖 / Wave**: Wave 3（依赖 T2）
* **回退方式**: 移除 Auth 中间件和 auth 路由注册
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A

#### Task 4: Web 登录页 + 路由守卫

* **目标**: 添加登录页面和导航守卫
* **不包含范围**: API client 修改（见 T5）
* **涉及文件**: `web/src/views/LoginView.vue`, `web/src/router/index.ts`
* **关键设计**:
  - 登录页：用户名/密码表单 → 调用 login API → 存储 token → 跳转
  - 路由守卫 `beforeEach`：无 token 时强制跳转 /login
  - /login 路由不触发守卫（白名单）
* **验收标准**:
  - 未登录访问 / → 跳转 /login
  - 输入凭据 → 登录成功 → 跳转 /
  - 登录失败 → 显示错误信息
  - 有 token 时访问 /login → 跳转 /
* **验证步骤**: 浏览器手动验证（L4）
* **测试要求**: 组合 T4 一起在浏览器中验证 L4 chain
* **依赖 / Wave**: Wave 4（依赖 T3）
* **回退方式**: 移除 /login 路由和守卫
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A（新页面，无基线）

#### Task 5: Web API 客户端认证适配

* **目标**: API 客户端自动携带 token，处理 401 和 refresh
* **不包含范围**: LoginView（见 T4）
* **涉及文件**: `web/src/api/client.ts`, `web/src/composables/useAuth.ts`
* **关键设计**:
  - `useAuth.ts` composable: `getToken()`, `setTokens()`, `clearTokens()`, `refreshAccessToken()`
  - Token 存储在 localStorage
  - apiRequest 添加 `Authorization: Bearer <token>` header
  - 401 时尝试 refresh token → 成功则重试请求 → 失败则清除 token 跳转登录
  - chatStream 函数通过 query param `?token=xxx` 传递 token
* **验收标准**: 所有 API 调用自动携带 token；Token 过期自动 refresh；Refresh 失败跳转登录
* **验证步骤**: 组合 T4 一起在浏览器中验证 L4 chain
* **测试要求**: 组合 T4 一起在浏览器中验证 L4 chain
* **依赖 / Wave**: Wave 4（依赖 T3）
* **回退方式**: 移除 auth header 注入逻辑
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A（新逻辑，无基线）

#### Task 6: 配置扩展

* **目标**: 添加 AuthConfig 到配置系统
* **不包含范围**: 无
* **涉及文件**: `internal/config/config.go`, `config.yaml`
* **关键设计**:
  ```go
  type AuthConfig struct {
      Username     string `yaml:"username"`
      Password     string `yaml:"password"`
      JWTSecret    string `yaml:"jwt_secret"`
      AccessExpiry time.Duration `yaml:"access_expiry"`
      RefreshExpiry time.Duration `yaml:"refresh_expiry"`
  }
  ```
* **验收标准**: config.yaml 新增 auth 段，环境变量覆盖生效
* **验证步骤**: `go test ./internal/config/...` (如有 config test)
* **测试要求**: go test ./internal/config/... (如有 config test)
* **依赖 / Wave**: Wave 1（无依赖）
* **回退方式**: 移除 AuthConfig 字段
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A

#### Task 7: E2E 认证场景验证

* **目标**: 编写 E2E curl 脚本覆盖认证全链路
* **不包含范围**: 无
* **涉及文件**: `test/e2e/curl/scenario-6-auth.sh`
* **验收标准**:
  - 场景 A: 无 token 访问 /api/sessions → 401
  - 场景 B: 登录 → access token → 访问 /api/sessions → 200
  - 场景 C: 过期 token → 401 → refresh → 新 token → 200
  - 场景 D: SSE stream 用 query param token → 200 + stream data
* **验证步骤**: `./test/e2e/curl/scenario-6-auth.sh` 全部 PASS
* **测试要求**: 4 个场景全部 PASS
* **依赖 / Wave**: Wave 5（依赖 T4+T5）
* **回退方式**: 删除脚本
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A（新增，无基线）

#### Spec 覆盖映射

| Spec 章节 / 映射编号 | 覆盖 Task | 说明 |
|----------------------|-----------|------|
| V1 (JWT 生成/验证) | T1 | package test |
| V2 (中间件拦截) | T2 | middleware test |
| V3 (登录全链路) | T3, T7 | handler + E2E |
| V4 (Web 登录页) | T4 | manual verification |
| V5 (SSE stream) | T5, T7 | query token + E2E |
| V6 (Token 过期+刷新) | T1, T5, T7 | package + client + E2E |
| F1 (login 端点) | T3 | handler |
| F2 (refresh 端点) | T3 | handler |
| F3 (Auth 中间件) | T2, T3 | middleware + integration |
| F4 (Web 登录页) | T4 | Vue component |
| F5 (API client) | T5 | client.ts |
