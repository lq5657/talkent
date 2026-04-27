# 会话记忆管理模式

> 来源: `chat-session` change 归档

## 滑动窗口 + LLM 摘要压缩

- 滑动窗口保留最近 N 条消息原文，窗口外消息通过 LLM 生成摘要注入 System Prompt
- 摘要作为 System Prompt 补充（`enrichedPrompt = systemPrompt + "\n\n## 对话历史摘要\n" + summary`），窗口内消息始终使用原文
- 摘要失败时降级为仅窗口模式（memory_source="window"），不阻断对话
- 摘要不持久化（MVP 阶段每轮重新生成），避免引入额外存储和一致性维护

## Service 层聚合视图

- Handler 不直接访问 Service 内部的 store 字段或做 JSON 反序列化
- Service 提供 `GetSessionDetail` 返回 Handler 需要的聚合数据（含计算字段如 currentRound/messageCount/roleDescription）
- Service 方法返回专用 Result 结构体（如 `EndSessionResult`）而非原始 store 模型，Handler 只关心展示逻辑

## 集中 JSON Unmarshal

- 当 store 中多个 JSON 字段需要反序列化（如 RoleConfig/Goals/Dimensions），集中到 `parseSessionConfig` 方法
- 每次调用都检查错误，避免静默零值传入业务逻辑
- 多个 Service 方法复用同一解析入口

## 测试模式

- `mockClient` 实现 `llm.Client` 接口，可配置 response/error/called 标志
- Manager 测试用 `makeMessages(n)` 工厂生成指定数量的测试消息
- Handler 测试使用 `http.ServeMux` + `httptest.NewRecorder` + `req.SetPathValue`
