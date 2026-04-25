---
change_id: llm-client
created: 2026-04-25
updated: 2026-04-25
---

### 任务拆分 — LLM 客户端抽象

#### 前置条件

* [x] `spec.md` 已确认且 `status = apply`
* [x] HARD-GATE 已通过
* [x] `scaffold-project` 已完成

#### Task 1: 添加 go-openai 依赖，定义 Client 接口与消息类型

* **目标** : 引入 `sashabaranov/go-openai`，在 `internal/llm/` 定义 `Client` 接口、`ChatMessage`/`ChatResponse` 等领域类型，与 go-openai 类型解耦
* **不包含范围** : 不实现 Client（属于 Task 2），不编写测试（属于 Task 4）
* **涉及文件** :
  * `internal/llm/client.go` — 新建，接口 + 领域类型
  * `go.mod` / `go.sum` — 修改，新增 `sashabaranov/go-openai` 依赖
* **关键签名** :
  ```go
  // internal/llm/client.go
  type MessageRole string
  const (
      RoleSystem    MessageRole = "system"
      RoleUser      MessageRole = "user"
      RoleAssistant MessageRole = "assistant"
  )

  type ChatMessage struct {
      Role    MessageRole
      Content string
  }

  type ChatResponse struct {
      Content     string
      Model       string
      PromptTokens   int
      CompletionTokens int
      TotalTokens     int
  }

  type ChatOptions struct {
      Temperature float32
      MaxTokens   int
  }

  type Client interface {
      Chat(ctx context.Context, messages []ChatMessage, opts *ChatOptions) (*ChatResponse, error)
  }
  ```
* **验收标准** : `go build ./...` 通过；`Client` 接口包含 `Chat` 方法；领域类型已定义，不暴露 go-openai 类型
* **验证步骤** : `go build ./...`（V1）
* **测试要求** : L2，V1 在本 task 关闭为 `apply-covered`
* **依赖 / Wave** : wave-1，无前置依赖
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/pre-apply.json -> baseline/post-task-1.json`

#### Task 2: 实现 go-openai Client 并集成配置

* **目标** : 实现 `internal/llm/openai.go`，使用 go-openai SDK 创建客户端，通过 `config.LLMConfig` 配置 BaseURL/APIKey/Model/Timeout；在 `cmd/server/main.go` 中初始化 LLM 客户端
* **不包含范围** : 不实现重试逻辑（属于 Task 3），不编写测试（属于 Task 4）
* **涉及文件** :
  * `internal/llm/openai.go` — 新建，go-openai 实现
  * `internal/llm/convert.go` — 新建，领域类型 ↔ go-openai 类型转换
  * `cmd/server/main.go` — 修改，新增 LLM 客户端初始化
* **关键签名** :
  ```go
  // internal/llm/openai.go
  type OpenAIClient struct { ... }
  func NewClient(cfg *config.LLMConfig, logger *slog.Logger) (*OpenAIClient, error)
  func (c *OpenAIClient) Chat(ctx context.Context, messages []ChatMessage, opts *ChatOptions) (*ChatResponse, error)

  // internal/llm/convert.go
  func toOpenAIMessages(messages []ChatMessage) []openai.ChatCompletionMessage
  func fromChatResponse(resp openai.ChatCompletionResponse) *ChatResponse
  ```
* **验收标准** : `go build ./...` 通过；`NewClient` 使用 `openai.DefaultConfig(cfg.APIKey)` + `config.BaseURL = cfg.BaseURL`；`Chat()` 正确调用 `CreateChatCompletion`；`cmd/server/main.go` 新增 LLM 客户端初始化
* **验证步骤** : `go build ./...`（V1, V3）
* **测试要求** : L2，V1 和 V3 在本 task 关闭为 `apply-covered`（编译级，完整功能验证在 Task 4）
* **依赖 / Wave** : wave-2，依赖 Task 1
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-task-1.json -> baseline/post-task-2.json`

#### Task 3: 错误处理、重试策略与日志

* **目标** : 在 `OpenAIClient.Chat()` 中实现错误分类处理（429/500 重试、401 不重试、其他记录）、日志记录（不含 API Key）和 context 超时控制
* **不包含范围** : 不编写测试（属于 Task 4），不修改接口定义
* **涉及文件** :
  * `internal/llm/openai.go` — 修改，错误处理、重试、日志增强
* **关键签名** :
  ```go
  // internal/llm/openai.go
  const maxRetries = 3

  func isRetryableError(err error) bool
  func (c *OpenAIClient) Chat(ctx context.Context, messages []ChatMessage, opts *ChatOptions) (*ChatResponse, error)
  // Chat 内部增加重试循环、错误分类、日志
  ```
* **验收标准** : 429/500 自动重试，最多 3 次，指数退避；401 直接返回不重试；日志记录 provider、model、消息数、token 用量、错误类型和 HTTP 状态码；日志不含 API Key；context 超时由 `config.LLMConfig.Timeout` 控制
* **验证步骤** : `go build ./...` 通过（V1）；完整功能验证在 Task 4
* **测试要求** : L2，V4/V5/V6 依赖 Task 4 单元测试闭环
* **依赖 / Wave** : wave-3，依赖 Task 2
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-task-2.json -> baseline/post-task-3.json`

#### Task 4: 单元测试

* **目标** : 为 `internal/llm/` 编写单元测试，使用 httptest 模拟 LLM API 端点，覆盖正常调用、错误处理、重试逻辑和日志脱敏
* **不包含范围** : 不测试真实 LLM API 调用（属于 L4 集成验证），不测试其他模块
* **涉及文件** :
  * `internal/llm/client_test.go` — 新建
  * `internal/llm/openai_test.go` — 新建
* **关键签名** :
  ```go
  // internal/llm/openai_test.go
  func TestChatCompletion_Success(t *testing.T)
  func TestChatCompletion_CustomBaseURL(t *testing.T)
  func TestChatCompletion_RetryOn429(t *testing.T)
  func TestChatCompletion_NoRetryOn401(t *testing.T)
  func TestChatCompletion_ContextCancellation(t *testing.T)
  func TestChatCompletion_APINotInLogs(t *testing.T)
  ```
* **验收标准** : `go test ./internal/llm/...` 全部通过；覆盖 V2-V6 所有验证项
* **验证步骤** : `go test ./internal/llm/... -v`（V2, V3, V4, V5, V6）
* **测试要求** : L2，V2-V6 在本 task 关闭为 `apply-covered`
* **依赖 / Wave** : wave-4，依赖 Task 3
* **回退方式** : git revert
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-task-3.json -> baseline/post-task-4.json`

#### 执行日志

| Task | 状态 | 涉及文件 | 验证证据 | 备注 |
|------|------|----------|----------|------|
| T1 | done | internal/llm/client.go, go.mod, go.sum | go build PASSED | go-openai v1.41.2 |
| T2+T3 | done | internal/llm/openai.go, internal/llm/convert.go, cmd/server/main.go | go build PASSED, go vet PASSED | Task 2+3 合并实现：重试是 Chat() 内禀逻辑 |
| T4 | done | internal/llm/openai_test.go | go test 19 passed | httptest 模拟，覆盖 V2-V6 |
