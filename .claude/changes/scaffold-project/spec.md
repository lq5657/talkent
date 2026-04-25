---
change_id: scaffold-project
status: review
depends_on: []
parallel_safe: true
branch: feat/scaffold-project
created: 2026-04-25
updated: 2026-04-25
complexity: simple
---

### 项目脚手架搭建 — Talkent

#### 0.1 需求收敛记录

`cc-new-project 项目定义` → `确定 Phase 0 首要 change` → `搭建 Talkent Go 项目工程基础`（收敛方式：`direct`）

#### 1. 背景与目标

Talkent 是一个角色扮演对话训练智能体，当前处于 Phase 0，需要在独立仓库中搭建 Go 项目工程基础。

本次 change 目标：建立项目骨架，包括 Go 模块、目录结构、配置管理、SQLite 初始化、HTTP Server 框架、结构化日志，确保 `go build ./...` 通过且服务可启动响应健康检查。

#### 1.0 路线图对齐

- 属于 Phase 0 的第一个 change，按 MVP 路线图 `scaffold-project` 执行
- 是所有后续 change 的前置依赖（`llm-client`、`role-and-goal`、`chat-session`、`analysis-engine` 均依赖本项目骨架）
- 与推荐 backlog 一致，未偏离

#### 1.1 本次不做

- LLM API 调用逻辑（属于 `llm-client`）
- 角色/目标/维度业务逻辑（属于 `role-and-goal`）
- 对话会话管理（属于 `chat-session`）
- 分析引擎（属于 `analysis-engine`）
- Web 前端（属于 `web-frontend`，Phase 1）
- Docker/部署配置
- 用户认证/授权
- 单元测试（scaffold 阶段仅有骨架，无业务逻辑可测；构建验证即足够）

#### 2. 代码现状（Research Findings）

绿地项目，当前仓库无 Talkent 业务代码。Harness 框架（`.claude/`）提供开发流程管理，不包含业务实现参考。

#### 3. 功能点

* [ ] **Go 模块初始化**：`go mod init`，声明模块路径
* [ ] **目录结构**：按架构草图创建 `cmd/`、`internal/`、`web/` 目录树
* [ ] **配置管理**：YAML 配置文件 + 环境变量覆盖，加载为 Config 结构体
* [ ] **SQLite 初始化**：启动时自动建表（sessions、messages、analysis_reports）
* [ ] **HTTP Server**：启动 HTTP 服务，提供 `/health` 端点
* [ ] **结构化日志**：统一 `log/slog` 初始化，含时间（微秒）、等级、源码位置
* [ ] **工程配套**：`.gitignore`、`Makefile`、`README.md` 骨架

#### 4. 业务规则

无特殊规则。纯工程脚手架，不涉及资金、状态流转或权限变更。

#### 5. 数据变更

- **是否涉及 migration**：否（自动建表，无历史数据迁移）
- **变更类型**：expand（新建表）
- **兼容窗口**：不适用（绿地项目）

| 操作 | 表名 | 字段/索引 | 说明 | 风险 |
|------|------|-----------|------|------|
| 新建 | sessions | id, role_config(json), goals(json), dimensions(json), status, round_limit, created_at, updated_at | 对话会话 | 低 |
| 新建 | messages | id, session_id, role(user/assistant), content, sequence_num, created_at | 对话消息 | 低 |
| 新建 | analysis_reports | id, session_id, dimension_results(json), created_at | 分析报告 | 低 |

#### 6. 接口变更

- **是否涉及对外契约变更**：否
- **兼容性分类**：无

| 操作 | 接口 | 方法 | 变更内容 | 兼容性 |
|------|------|------|----------|--------|
| 新增 | `/health` | GET | 返回 `{"status":"ok"}` | 新增 |

#### 7. 影响范围

纯新增文件，无存量代码影响。

#### 7.1 配置变更

- **是否涉及配置项或环境变量变更**：是
- **配置来源**：YAML 配置文件 + 环境变量覆盖
- **新增/变更配置项**：

| 配置项 | 默认值 | 必填 | 说明 |
|--------|--------|------|------|
| `server.port` | `8080` | 否 | HTTP 监听端口 |
| `server.host` | `0.0.0.0` | 否 | 监听地址 |
| `database.path` | `./talkent.db` | 否 | SQLite 文件路径 |
| `log.level` | `info` | 否 | 日志级别 |
| `log.file` | `""` (stdout) | 否 | 日志文件路径 |
| `llm.provider` | - | 是 | LLM 提供商标识 |
| `llm.base_url` | - | 是 | LLM API 地址 |
| `llm.api_key` | - | 是 | LLM API Key（建议通过环境变量 `TALKENT_LLM_API_KEY` 注入） |
| `llm.model` | - | 是 | 模型名称 |

- **默认值与是否安全**：`server.port` 默认 8080 安全；`llm.api_key` 无默认值，必须注入
- **环境差异**：开发/生产通过不同配置文件区分

#### 8. 风险与关注点

无额外风险。纯工程脚手架，不涉及资金、外部依赖、状态流转。

#### 8.1 日志与可观测性

- **是否新增运行时日志点**：是（启动日志、请求日志）
- **涉及入口**：HTTP Server 启动、`/health` 请求处理
- **使用的 logger**：`log/slog`
- **关键字段**：无（MVP 阶段暂无 request_id/trace_id）
- **日志落点**：默认 stdout；可通过 `log.file` 配置落盘
- **日志格式**：时间（微秒）、等级、消息、文件名:行号、函数/方法名

#### 9. 测试策略

- **测试范围**：`go build ./...` 构建验证
- **最低验证等级**：L1（Build）
- **验证证据要求**：全量构建通过
- **若无法达到目标等级的替代方案**：不适用（L1 为最低等级）

#### 9.1 需求-验证映射

| 编号 | 需求项 / 风险点 | 最低验证等级 | 证据类型 | 建议验证动作 | 对应 Task | 闭环状态 |
|------|------------------|--------------|----------|--------------|-----------|----------|
| V1 | Go 模块构建通过 | L1 | build | `go build ./...` | Task 1 | apply-covered |
| V2 | 配置加载不报错 | L1 | build | 启动服务，检查日志 | Task 2 | apply-covered |
| V3 | SQLite 自动建表 | L1 | build | 启动后检查 talkent.db 文件生成 | Task 3 | apply-covered |
| V4 | HTTP 健康检查 | L1 | build | `curl /health` 返回 200 | Task 4 | apply-covered |
| V5 | 日志格式化输出 | L1 | build | 检查启动日志格式 | Task 5 | apply-covered |
| V6 | 工程配套文件 | L1 | doc-check | `.gitignore`、`Makefile`、`README.md` 存在 | Task 1 | apply-covered |

#### 9.2 发布与回滚

低风险脚手架，直接代码提交 + git revert 即可回滚。

- **发布方式**：直接提交
- **回滚路径**：代码回滚（git revert）
- **发布后观察窗口**：不适用

#### 10. 待澄清

- [x] Go 模块路径：`github.com/lq5657/talkent`，代码放在当前目录下

#### 11. 方案比较

不触发：绿地项目，按架构草图中已确认的技术方向执行。

#### 12. 技术决策

| 决策 | 选择 | 放弃的方案 | 原因 |
|------|------|-----------|------|
| HTTP 路由 | `net/http` 标准库 `ServeMux`（Go 1.22+ 路径参数支持） | `chi`、`gin` | MVP 接口简单，标准库足够，零外部依赖 |
| SQLite 驱动 | `modernc.org/sqlite`（纯 Go，无 CGO） | `mattn/go-sqlite3`（需要 CGO） | 跨平台编译简单，无 CGO 依赖 |
| 配置格式 | YAML（`gopkg.in/yaml.v3`） | JSON、TOML | 可读性好，Go 生态主流 |
| 配置加载策略 | 配置文件 + 环境变量覆盖（`TALKENT_` 前缀） | 仅配置文件 / 仅环境变量 | 兼顾便捷和安全性（API Key 必须走环境变量） |
| 日志 | `log/slog` + `AddSource` | `zap`、`zerolog` | 标准库，零外部依赖，满足 MVP 需求 |

#### 13. 执行日志

| T1 | done | go.mod, cmd/server/main.go, .gitignore, Makefile, README.md | pre-apply.json -> post-apply.json | commit `a112137` |
| T2 | done | internal/config/config.go, config.example.yaml, go.sum | post-apply.json | commit `2bd2ef1` |
| T3 | done | internal/store/db.go, internal/store/schema.go, go.sum, go.mod | post-apply.json | commit `2d103c9` |
| T4 | done | internal/server/server.go, cmd/server/main.go | post-apply.json | commit `fa02f4c` |
| T5 | done | internal/log/log.go, cmd/server/main.go | post-apply.json | commit `fa02f4c` |
| T6 | done | cmd/server/main.go (full wiring) | post-apply.json (go build + go vet + E2E health check PASSED) | commit `fa02f4c` |

#### 14. 审查结论

（待 cc-review）

#### 15. 确认记录（HARD-GATE）

* **confirmed_at**：2026-04-25
* **confirmed_by**：lq5657
* **confirmed_spec_revision**：`spec.md` @ 2026-04-25，6 个功能点，V1-V6 验证映射，模块路径 `github.com/lq5657/talkent`
* **confirmed_tasks_revision**：`tasks.md` @ 2026-04-25，6 个 task，4 个 wave
* **confirmed_scope**：Go 模块初始化、目录、配置、SQLite、HTTP Server、日志；不含业务逻辑和前端
* **accepted_risks**：无额外风险
* **human_review_required**：false
* **human_review_status**：not_required
