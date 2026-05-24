---
id: sse-eventsource-auth-header
type: pitfall
status: confirmed
applies_to:
  - cc-propose
  - cc-apply
  - cc-review
triggers:
  - SSE
  - EventSource
  - stream
  - token
  - auth
  - 流式
confidence: confirmed
evidence:
  - .cc/changes/android-backend-ready/log.md
  - internal/auth/middleware.go:extractToken
  - web/src/api/client.ts:chatStream
---

# SSE/EventSource 不支持自定义 HTTP Header

## Rule / Insight

浏览器 `EventSource` API 和 `fetch` ReadableStream 的 SSE 模式下，**无法设置自定义 HTTP header**（包括 `Authorization`）。认证 token 必须通过 URL query param `?token=xxx` 传递。后端 Auth 中间件需同时支持从 `Authorization: Bearer <token>` header 和 `?token=<token>` query param 提取 token。

## Applies When

- 新增或修改 SSE/EventSource 流式端点
- Web 前端使用 `EventSource` 或 `fetch` + `ReadableStream` 消费流式 API
- 需要在流式端点传递认证凭据

## Does Not Apply When

- 非浏览器客户端（Android/iOS 原生、curl）可以直接使用自定义 header
- 非 SSE 的普通 HTTP API 调用

## Evidence

- `android-backend-ready` change: E2E scenario-6-D (SSE stream with query param token: PASS)
- 技术决策记录: spec.md §12 — 选择 query param 而非自定义 header

## Usage Notes

- 中间件应优先检查 `Authorization` header，fallback 到 `?token` query param
- Query param 中的 token 可能被反向代理/网关日志记录——这是已知 trade-off，非敏感环境下可接受
- Android/iOS 原生客户端可继续使用 `Authorization` header
