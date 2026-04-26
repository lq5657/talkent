# 角色模块集成模式

> 来源: `role-and-goal` change 归档

## 模板匹配+LLM 兜底模式

- 推荐链路使用二级策略：先查表/模板匹配，匹配不到再 LLM 生成
- Service 层封装策略选择，Handler 层只调 Service，不感知策略细节
- 返回值携带 `source` 字段（"template"/"table"/"llm"），便于客户端区分来源
- LLM Prompt 使用硬编码 System Prompt 常量 + 用户输入仅出现在 User Prompt，防止注入

## server.New 回调路由注册

- `server.New(cfg, db, logger, registerRoutes func(*http.ServeMux))` 接收回调
- 各模块 Handler 实现 `RegisterRoutes(mux *http.ServeMux)` 方法
- main.go 依次注入：`roleHandler.RegisterRoutes` → `server.New(..., roleHandler.RegisterRoutes)`
- 新增模块无需修改 server.go，只需在 main.go 追加注入

## 测试模式

- `mockLLMClient`：实现 `llm.Client` 接口，可配置 response/error，用于 Service 和 Handler 单元测试
- `capturingClient`：包装 mockLLMClient，拦截 `ChatMessage` 切片，用于验证 Prompt 注入防护（System Prompt 不被用户输入污染）
- Handler 测试使用 `httptest.NewRecorder` + `http.ServeMux`，与 llm-client 的 `httptest.NewServer` 模式互补
