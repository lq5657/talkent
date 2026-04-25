# Dev Map

本文件记录项目开发导航图，服务于 `cc-propose`、`cc-apply`、`cc-review` 和新成员理解代码边界。
它只沉淀可复用、可验证、相对稳定的工程事实；临时猜测必须写入"待确认"，不得写成事实。

```text
last_updated: 2026-04-25
updated_by: cc-new-project
confidence: draft
```

## 1. 模块导航

| 模块 / 边界 | 主要路径 | 职责 | 入口 | 关键依赖 | 信心 |
|-------------|----------|------|------|----------|------|
| HTTP Server | `cmd/server/` | 路由、请求解析、响应 | `main.go` | 配置管理 | medium |
| 配置管理 | `internal/config/` | 多模型配置加载与校验 | `config.Load()` | 无 | medium |
| LLM 客户端 | `internal/llm/` | OpenAI 兼容协议适配 | `llm.Client` 接口 | 配置 | medium |
| 角色与目标 | `internal/role/` | 角色定义、目标推荐、维度映射 | `role.Service` | LLM 客户端 | medium |
| 会话管理 | `internal/session/` | 对话生命周期编排 | `session.Service` | LLM、role、memory | medium |
| 对话记忆 | `internal/memory/` | 滑动窗口、摘要压缩 | `memory.Manager` | 无 | medium |
| 分析引擎 | `internal/analysis/` | 维度分析、报告生成 | `analysis.Engine` | LLM、role、session | medium |
| 数据持久化 | `internal/store/` | SQLite 读写 | `store.Store` 接口 | 无 | medium |
| Web 前端 | `web/` | 设定页、对话页、报告页 | Phase 1 | HTTP API | low |

## 2. 关键链路

| 链路 | 入口 | 主要步骤 | 影响数据 | 常用验证 | 风险 |
|------|------|----------|----------|----------|------|
| 角色设定 → 目标推荐 | `POST /api/roles/define` | 用户提交角色描述 → LLM 推荐目标 → 返回目标列表 | Role, TrainingGoal | API 测试 | LLM 推荐质量 |
| 对话主流程 | `POST /api/sessions/{id}/chat` | 角色注入 → LLM 生成回复 → 记忆更新 → 轮数检查 | Session, Message | API 测试 + LLM 输出检查 | token 超限、API 失败 |
| 分析报告生成 | `POST /api/sessions/{id}/analyze` | 提取对话 → 构造分析 Prompt → LLM 分析 → JSON 解析 → MD 渲染 | AnalysisReport | API 测试 + 人工评估报告质量 | 分析质量、JSON 解析失败 |
| 模型配置切换 | 启动加载 | 配置文件 → Config 结构体 → LLM Client | LLMConfig | 启动测试 | 配置错误 |

## 3. 测试与验证入口

| 范围 | 测试路径 / 命令 | 覆盖对象 | 适用场景 | 备注 |
|------|----------------|----------|----------|------|
| 全量构建 | `go build ./...` | 所有包 | L1 构建验证 | Phase 0 即要求 |
| LLM 客户端 | `go test ./internal/llm/...` | LLM 适配层 | 单独调试 LLM 接入 | 可能需 mock LLM API |
| 分析引擎 | `go test ./internal/analysis/...` | 分析逻辑 | 验证 Prompt 和报告生成 | 关键质量模块 |
| E2E 手工 | curl / 浏览器 | 全链路 | Phase 0 手工验证、Phase 1 浏览器验证 | |

## 4. 易错边界

| 边界 | 证据位置 | 影响 | 处理原则 | 信心 |
|------|----------|------|----------|------|
| LLM API Key 泄露 | 配置文件、环境变量 | 安全风险 | 只用环境变量或外部配置注入，不提交到仓库 | high |
| LLM 返回格式不稳定 | `analysis.Engine` | 分析报告生成失败 | JSON Schema 约束 + 解析容错 + 重试 | medium |
| 长对话 token 超限 | `memory.Manager` | 对话中断或质量下降 | 滑动窗口 + 对话长度预估 | medium |
| 角色 Prompt 注入 | `role.Service` | AI 角色设定被用户输入覆盖 | System Prompt 与 User Prompt 严格分离 | medium |
| 跨模型兼容性 | `llm.Client` | 切换模型后行为异常 | OpenAI 兼容协议标准化 + 多模型测试 | low |

## 5. Change 影响索引

| change_id | 影响模块 | 影响链路 | 关联验证 | 状态 |
|-----------|----------|----------|----------|------|
| 待创建 | - | - | - | - |

## 6. 待确认事项

| 问题 | 影响范围 | 建议确认方式 | 优先级 |
|------|----------|--------------|--------|
| Go HTTP 路由库选择 | HTTP Server | 评估 `net/http` vs `chi` vs `gin` | P1 |
| SQLite 纯 Go 驱动选型 | 数据持久化 | 评估 `modernc.org/sqlite` vs `zombiezen.com/go/sqlite` | P1 |
| 前端框架选型 | Web 前端 | Phase 0 完成后再定 | P2 |
| 项目仓库名称和位置 | 全项目 | 用户决定 | P1 |

## 更新规则

- `cc-init` 只写基础导航：模块候选、入口、测试入口和待确认事项。
- `cc-enrich-context` 负责补关键链路、易错边界、验证入口和信心等级。
- `cc-apply` 只有在代码结构、模块边界或验证入口发生实质变化时才更新本文件。
- 不得写入无证据的架构判断、敏感信息、临时调试日志或单次执行细节。
