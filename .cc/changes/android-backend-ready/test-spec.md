---
change_id: android-backend-ready
status: propose
created: 2026-05-24
mode: supplement
---

### 测试 Spec — 后端 Android 适配

#### 1. 测试框架

| 项目 | 值 |
|------|----|
| 测试执行器 | go test (Go), curl + bash (E2E) |
| 增强断言/Mock框架 | httptest (stdlib) |
| 已有测试数量 | 13 (auth package) + 原有 115 (other packages) = 128 |
| 已有测试风格 | 表驱动测试 + httptest |

#### 1.1 测试层级选择

| 项目 | 值 |
|------|----|
| 主测试层级 | chain (L3) + manual (L4) |
| 选择原因 | V3/V5 需要全链路 chain 证据（login → token → API），V4 需要浏览器手动验证 |
| `cc-test` 模式 | supplement |
| 更高层测试是否跳过 | 否 |
| 跳过原因 | N/A |

#### 1.2 已完成的最小验证 (cc-apply L2)

| Task | 涉及映射项 | cc-apply 内已完成验证 | 证据 | 尚未覆盖的风险 |
|------|------------|----------------------|------|---------------|
| T1 | V1,V6 | JWT 生成/验证/过期/刷新 | jwt_test.go 6 pass | -- |
| T2 | V2 | 中间件 7 场景 | middleware_test.go 7 pass | -- |
| T3 | V3 | handler 集成 | go build + 128 tests pass | chain (login→token→API) |
| T4 | V4 | 前端构建 | npm run build pass | manual browser login |
| T5 | V5 | API client 构建 | npm run build pass | SSE stream with real token |
| T6 | -- | config 加载 | go build pass | -- |
| T7 | V3,V5 | E2E 脚本已编写 | 脚本已存在 | 脚本未实际运行 |

#### 1.3 验证差距与补强计划

| 映射编号 | 需求项 / 风险点 | 当前状态 | cc-apply 证据 | 本次补强 | 跳过原因 | 替代证据 | 剩余风险 |
|----------|------------------|----------|-------------|----------|----------|----------|----------|
| V3 | 登录全链路 (L3) | apply-covered | L2 package tests | E2E scenario-6-A+B 运行 | -- | -- | -- |
| V4 | Web 登录页 (L4) | apply-covered | npm build pass | 浏览器手工验证或截图 | 无浏览器环境时暂跳过 | -- | 仅 build 通过 |
| V5 | SSE stream (L3) | apply-covered | L2 middleware tests | E2E scenario-6-D 运行 | -- | -- | -- |
| V6 | Token 刷新 (L2) | apply-covered | jwt_test.go | E2E scenario-6-C 运行 | -- | -- | -- |

#### 2. 覆盖范围

##### P0 — 核心认证链路 (chain)

| 方法 | 场景 | 输入 | 预期结果 |
|------|------|------|----------|
| POST /api/auth/login | 正确凭据 | {"username":"admin","password":"admin"} | 200 + access_token + refresh_token |
| GET /api/sessions | Bearer token | Authorization: Bearer <token> | 200 |
| GET /api/sessions | 无 token | 无 header | 401 |
| POST /api/auth/refresh | 有效 refresh_token | {"refresh_token":"<token>"} | 200 + 新 access_token |

#### 2.1 分层覆盖说明

| 层级 | 是否覆盖 | 覆盖对象 | 说明 |
|------|----------|----------|------|
| unit | ✅ | JWT service, middleware | 13 tests, L2 达标 |
| chain | ✅ (本次补强) | login → token → API → 401 | E2E scenario-6 |
| manual | 部分 (本次) | Web login 页面 | 前端构建通过，浏览器验证由用户自行完成 |

##### 不测试

- 无

#### 3. 执行计划

* [x] Step 1: Write test-spec.md
* [x] Step 2: Start server `go run ./cmd/server/ &` → server started on 0.0.0.0:8080
* [x] Step 3: Run `./test/e2e/curl/scenario-6-auth.sh` → 6/6 PASS
* [x] Step 4: Capture output → all scenarios pass (A:401, B:login+API, C:refresh, D:SSE stream)
* [x] Step 5: Sync spec.md validation mapping → V3/V5/V6 test-covered, V4 apply-covered
* [x] Step 6: Server stopped

#### 4. 执行证据

```
=== Scenario 6: JWT Authentication ===

A: No token → 401           → PASS (HTTP 401)
B: Login → token → API      → PASS (HTTP 200 ×2)
D: SSE stream query token   → PASS (token accepted)
C: Refresh → new token      → PASS (HTTP 200 + new token)

Results: 6 passed, 0 failed
```
