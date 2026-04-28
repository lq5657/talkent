# Task Board

本文件记录 change 级工作看板，用于快速判断当前仓库有哪些候选、进行中、阻塞或待归档的工作。
它不是 `spec.md` / `tasks.md` 的替代品；真实需求、验收和证据仍以单个 `changes/<change-id>/` 为准。

```text
last_updated: 2026-04-28
updated_by: cc-apply
```

## 1. 正式 Change

| change_id | 状态 | 来源 | 目标摘要 | 影响模块 | 阻塞 / 依赖 | 下一命令 | 最近证据 |
|-----------|------|------|----------|----------|-------------|----------|----------|
| scaffold-project | done | cc-archive | 项目脚手架：Go 模块、目录结构、配置、SQLite、HTTP Server | cmd/server, internal/{config,store,server,log} | 无 | - | review PASSED, Stage1+2 pass, 0 findings, 已归档 |
| llm-client | done | cc-archive | LLM 客户端抽象：go-openai SDK、单活跃 provider、错误重试 | internal/llm, cmd/server | scaffold-project(done) | - | review PASSED, F4 fixed, 21 tests, 已归档 |
| role-and-goal | done | cc-archive | 角色设定→目标推荐→维度确定的推导链路 | internal/role, internal/server, cmd/server | scaffold-project(done), llm-client(done) | - | review PASSED, 0C/0I/2Minor, 已归档, knowledge沉淀3条 |
| chat-session | done | cc-archive | 对话会话生命周期：创建、对话、记忆管理、结束 | internal/{session,memory,store}, cmd/server, internal/config | scaffold-project(done), llm-client(done), role-and-goal(done) | - | review PASSED, 0C/0I(2fixed)/0M(2fixed+1accepted), 23 tests, knowledge沉淀4条, 已归档 |
| analysis-engine | done | cc-archive | 多维度分析引擎：结构化分析+JSON解析容错+Markdown报告+自动触发 | internal/{analysis,store,session}, cmd/server, internal/config | scaffold-project(done), llm-client(done), role-and-goal(done), chat-session(done) | - | review PASSED, F1-F5 fixed, 95 tests, knowledge沉淀4条, 已归档 |

## 2. Backlog 候选

| 候选项 | 来源 | 推荐 change_id | 价值 | 前置条件 | 建议下一步 |
|--------|------|----------------|------|----------|------------|
| 项目脚手架搭建 | cc-new-project | `scaffold-project` | 工程基础：模块初始化、目录结构、配置、SQLite | 无 | `cc-propose scaffold-project` |
| LLM 客户端抽象 | cc-new-project | `llm-client` | 统一 LLM 接入，多模型可配置 | `scaffold-project` | 已提案为正式 change |
| 角色设定与目标推荐 | cc-new-project | `role-and-goal` | 角色→目标→维度的推导链路 | `scaffold-project` | 可与 LLM 客户端并行提案 |
| 对话会话管理 | cc-new-project | `chat-session` | 对话生命周期、角色注入、记忆管理 | `llm-client` | LLM 客户端完成后提案 |
| 分析引擎 | cc-new-project | `analysis-engine` | 多维度分析报告生成，核心差异化 | `llm-client`, `role-and-goal`, `chat-session` | 依赖就绪后提案 |
| Web 前端 | cc-new-project | `web-frontend` | 三页面 Web 应用（设定/对话/报告） | Phase 0 全部完成 | Phase 1 开始时提案 |
| E2E 集成 | cc-new-project | `e2e-integration` | 全链路联调、错误处理、边界覆盖 | Phase 0 + `web-frontend` | Web 前端完成后提案 |

## 3. 阻塞项

| change_id / 候选项 | 阻塞原因 | 需要谁确认 | 恢复入口 | 记录位置 |
|--------------------|----------|------------|----------|----------|
| 暂无 | - | - | - | - |

## 更新规则

- `cc-new-project` 可写入 Backlog 候选，但不得创建正式 change。
- `cc-propose` 创建正式 change 后，必须新增或更新正式 Change 行。
- `cc-apply`、`cc-test`、`cc-review`、`cc-fix` 和 `cc-archive` 必须同步状态、阻塞项和下一命令。
- 看板只保存摘要和导航，不复制 spec/tasks/review 的完整正文。
