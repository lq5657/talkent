---
name: SSE Flusher Middleware Pitfall
description: Middleware ResponseWriter wrappers must expose http.Flusher or SSE breaks
type: pitfall
status: confirmed
confidence: high
source: voice-interaction change (L4 verification)
---

**问题**: HTTP middleware 包装 `http.ResponseWriter` 时（如 timeout/deadline middleware），如果包装结构体没有暴露 `Flush()` 方法，SSE handler 中的 `w.(http.Flusher)` 类型断言会失败，导致 SSE 永远返回 "streaming not supported"。

**根因**: Go 的 `http.ResponseWriter` 接口不包含 `Flusher`，但标准库 `net/http` 的实现总是支持 `http.Flusher`。当自定义 wrapper struct 嵌入 `http.ResponseWriter` 时，embedding 并不会自动提升 `Flusher` 方法（因为 `http.ResponseWriter` 接口本身不声明 `Flush()`）。

**解决方案**: 在 custom `ResponseWriter` wrapper 上添加 `Flush()` 委托方法：

```go
func (tw *timeoutResponseWriter) Flush() {
    if f, ok := tw.ResponseWriter.(http.Flusher); ok {
        f.Flush()
    }
}
```

**适用场景**: 任何涉及 SSE、WebSocket 或长时间流式响应的 Go HTTP 中间件。
