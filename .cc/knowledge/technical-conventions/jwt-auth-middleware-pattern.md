---
id: jwt-auth-middleware-pattern
type: technical_convention
status: candidate
applies_to:
  - cc-propose
  - cc-apply
  - cc-review
triggers:
  - JWT
  - auth
  - login
  - token
  - refresh
  - 认证
confidence: candidate
evidence:
  - .cc/changes/android-backend-ready/
  - internal/auth/jwt.go
  - internal/auth/middleware.go
  - internal/auth/handler.go
---

# Go JWT 认证中间件模式

## Rule / Insight

Go HTTP 服务的 JWT 认证推荐模式：access token (HMAC-SHA256, 1h) + refresh token (7d)，中间件位于 CORS 之后/mux 之前，login/refresh 端点白名单，`/health` 和 OPTIONS 请求跳过认证。Token 提取优先 `Authorization: Bearer <token>` header，fallback 到 `?token=<token>` query param（SSE 兼容）。

## Applies When

- 需要为 Go HTTP API 添加认证
- 单用户/小规模认证场景（凭据来自配置文件，非数据库）

## Does Not Apply When

- 多用户注册/登录系统（需要数据库存储用户）
- OAuth2/第三方登录场景
- 需要 JWT 撤销/黑名单机制

## Evidence

- `android-backend-ready` change: 13 auth tests pass, E2E scenario-6 6/6 PASS
- `internal/auth/` 包完整实现: JWT service + middleware + handler

## Usage Notes

- `golang-jwt/jwt/v5` 是推荐的 JWT 库
- JWT secret 默认值应为占位符，生产环境通过环境变量注入
- 登录接口建议添加 per-IP 频率限制（如 5次/分钟）防止暴力破解
- 本模式为单用户设计；多用户场景需要扩展 credential store
