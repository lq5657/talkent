---
name: Go fmt %q vs %s JSON Pitfall
description: fmt.Fprintf with %s in hand-crafted JSON produces invalid output for special characters
type: pitfall
status: confirmed
confidence: high
source: voice-interaction change (F1 review finding)
---

**问题**: 在 `fmt.Fprintf` 中手工构建 JSON 字符串时，`%s` 直接嵌入字符串不做任何转义。如果字符串含 `"`、`\`、换行符等特殊字符，产生的 JSON 无效。

**错误示例**:
```go
// BAD: %s 不转义 — JSON 可能无效
fmt.Fprintf(w, "data: {\"error\":\"%s\"}\n\n", errMsg)
```

**正确做法**:
```go
// GOOD: %q 安全转义双引号和特殊字符
fmt.Fprintf(w, "data: {\"error\":%q}\n\n", errMsg)
```

**Go `%q` 的转义**: 产生 Go 语法安全加引号的字符串，转义 `"` → `\"`、`\` → `\\`、换行 → `\n` 等。这些转义与 JSON 标准兼容。

**更严格的做法**: 对于复杂 JSON，使用 `json.Marshal` 或 `json.NewEncoder`。

**适用场景**: 任何在 SSE、日志、模板中手工构建 JSON 字符串的地方。
