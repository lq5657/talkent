# Log — llm-client

## 2026-04-25 cc-propose

### 决策记录

| 决策 | 选择 | 理由 |
|------|------|------|
| 多模型策略 | 单活跃 provider | MVP 最简实现，用户选择单活跃 provider，切换靠改配置重启 |
| HTTP 客户端 | sashabaranov/go-openai | 用户明确选择；社区维护、类型安全、支持自定义 BaseURL |
| 接口与实现分离 | 领域类型 + Client 接口 | 不在接口层暴露 go-openai 类型，便于后续替换实现或 mock 测试 |

### Source-Driven Development 证据

| 外部依赖 | 查询来源 | 关键发现 | 信心 |
|----------|----------|----------|------|
| sashabaranov/go-openai | Context7 文档查询 | `DefaultConfig(apiKey)` + `config.BaseURL` 可适配非 OpenAI 端点；`APIError` 类型支持按状态码分类处理 | high |

### 待确认事项

- 无阻塞级待确认项

## 2026-04-25 cc-apply

### 执行过程

| 步骤 | 内容 | 结果 |
|------|------|------|
| pre-apply baseline | `go build ./...` + `go vet ./...` | PASSED |
| Task 1 | 添加 go-openai v1.41.2 依赖，定义 Client 接口与领域类型 | done |
| Task 2+3 | 实现 OpenAIClient + 错误处理/重试/日志（合并实现） | done |
| Task 4 | httptest 单元测试 19 passed | done |
| post-apply | `go build ./...` + `go vet ./...` + `go test ./internal/llm/...` | ALL PASSED |

### 实现决策

| 决策 | 选择 | 理由 |
|------|------|------|
| Task 2+3 合并 | 重试、日志、错误处理直接写在 `Chat()` 方法内 | 重试是调用流程内禀逻辑，拆出独立文件会导致中间态不可用 |
| 重试策略 | for 循环 + 指数退避 + context 取消检查 | 比 recursive 更直观，退避期间仍响应 context 取消 |
| `_ = llmClient` | main.go 中暂不传入 server | LLM 客户端将在 chat-session change 中 wire 到 server |

### 文件清单

| 文件 | 操作 | 行数 |
|------|------|------|
| `internal/llm/client.go` | 新建 | 35 |
| `internal/llm/openai.go` | 新建 | 107 |
| `internal/llm/convert.go` | 新建 | 31 |
| `internal/llm/openai_test.go` | 新建 | ~270 |
| `cmd/server/main.go` | 修改 | +6 |
| `go.mod` | 修改 | +1 dep |
| `go.sum` | 修改 | +indirect deps |

## 2026-04-26 cc-fix

### 修复记录

| Finding | 根因 | 最小修复 | Guard | 验证 |
|---------|------|----------|-------|------|
| F4 (Important) | NewClient 未将 cfg.Timeout 应用到 HTTP client | openai.go 新增 `if cfg.Timeout > 0 { ocfg.HTTPClient = &http.Client{Timeout: cfg.Timeout} }` | TestNewClient_ConfigTimeoutApplied + TestChatCompletion_HTTPTimesOutWithConfigTimeout | 21 tests passed, go build + vet PASSED |
