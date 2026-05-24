---
id: okhttp-interceptor-token-refresh
type: technical_convention
status: candidate
applies_to:
  - cc-propose
  - cc-apply
triggers:
  - OkHttp
  - Interceptor
  - token
  - refresh
  - 401
  - retry
confidence: candidate
evidence:
  - .cc/changes/android-core-shell/
  - android/.../AuthInterceptor.kt
---

# OkHttp Interceptor Token Auto-Refresh 模式

## Rule / Insight

OkHttp `Interceptor` 中拦截 401 响应 → 使用 refresh token 同步调用 refresh endpoint → 更新 access token → 重试原请求。Refresh 也失败时触发 logout event（`StateFlow<Boolean>` 由 UI 层观察跳转登录页）。

## Pattern

```kotlin
class AuthInterceptor(
    private val tokenManager: TokenManager,
    private val baseUrlProvider: () -> String
) : Interceptor {

    private val _logoutEvent = MutableStateFlow(false)
    val logoutEvent: StateFlow<Boolean> = _logoutEvent

    override fun intercept(chain: Interceptor.Chain): Response {
        // 1. Skip public endpoints
        // 2. Add Bearer token to request
        // 3. Execute request
        // 4. If 401 -> refreshAccessToken() sync
        // 5. If refresh OK -> retry request with new token
        // 6. If refresh fails -> emit logoutEvent, clear tokens
    }
}
```

## Applies When

- Android 客户端使用 OkHttp + Retrofit
- 后端使用 JWT access/refresh token 方案

## Does Not Apply When

- 使用 Ktor 或其他 HTTP 客户端
- Token 过期通过其他机制处理（如定时器预刷新）

## Evidence

- `android-core-shell` change: AuthInterceptor.kt 完整实现

## Usage Notes

- Interceptor 中 refresh 调用必须是同步的（阻塞 interceptor chain）
- 注意避免多个并发请求同时触发 refresh（可加 `synchronized` 或 `AtomicBoolean` 锁优化）
- `logoutEvent` 使用 `StateFlow` 而非 `Channel` 以便 UI 层 collect 时获取当前状态
- 当前实现为基础版，高并发场景建议加 refresh 锁
