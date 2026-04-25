---
change_id: scaffold-project
created: 2026-04-25
updated: 2026-04-25
---

### 任务拆分 — 项目脚手架搭建

#### 前置条件

* [ ] `spec.md` 已确认且 `status = propose`
* [ ] HARD-GATE 已通过
* [ ] Go 模块路径已确定

#### Task 1: Go 模块初始化与目录结构

* **目标** : `go mod init` 创建模块，建立完整目录树，编写 `.gitignore`、`Makefile`、`README.md`
* **不包含范围** : 不实现任何业务逻辑，不创建具体 Go 源文件（仅 `main.go` 的 `package main` 声明）
* **涉及文件** :
  * `go.mod` — 新建
  * `cmd/server/main.go` — 新建，仅 `package main` + `func main() {}`
  * `internal/config/` — 新建目录
  * `internal/llm/` — 新建目录
  * `internal/role/` — 新建目录
  * `internal/session/` — 新建目录
  * `internal/memory/` — 新建目录
  * `internal/analysis/` — 新建目录
  * `internal/store/` — 新建目录
  * `web/` — 新建目录
  * `.gitignore` — 新建
  * `Makefile` — 新建
  * `README.md` — 新建
* **关键签名** :
  ```go
  // cmd/server/main.go
  package main
  func main() {}
  ```
* **验收标准** : `go build ./...` 成功（即使 main 为空）；目录结构与架构草图一致；`.gitignore` 包含 `talkent.db`、`*.log`、`config.yaml`（本地配置）
* **验证步骤** : `go build ./...`（V1）；目视检查目录结构完整性（V6）
* **测试要求** : L1 build，不需要 TDD；V1 和 V6 在本 task 关闭为 `apply-covered`
* **依赖 / Wave** : wave-1，无前置依赖
* **回退方式** : 删除仓库目录即可（未提交 git 历史外的任何状态）
* **完成后状态** : done
* **并发注意事项** : 无（wave-1 起始 task）
* **Baseline / Delta** : `baseline/pre-apply.json -> baseline/post-task-1.json`
* **对应 commit** : `a112137`

#### Task 2: 配置管理

* **目标** : 实现 Config 结构体、YAML 配置加载、环境变量覆盖（`TALKENT_` 前缀）
* **不包含范围** : LLM 客户端实例化（仅加载配置，不连接），不处理配置热更新
* **涉及文件** :
  * `internal/config/config.go` — 新建，Config 结构体定义 + Load 函数
  * `internal/config/config_test.go` — 新建，Load 函数基础测试（可选，scaffold 阶段非必须）
  * `config.example.yaml` — 新建，配置模板
* **关键签名** :
  ```go
  // internal/config/config.go
  type Config struct {
      Server ServerConfig  `yaml:"server"`
      Database DBConfig    `yaml:"database"`
      Log LogConfig        `yaml:"log"`
      LLM LLMConfig        `yaml:"llm"`
  }

  type LLMConfig struct {
      Provider string `yaml:"provider"`
      BaseURL  string `yaml:"base_url"`
      APIKey   string `yaml:"api_key"`
      Model    string `yaml:"model"`
      Timeout  time.Duration `yaml:"timeout"`
  }

  // 新增
  func Load(path string) (*Config, error)
  ```
* **验收标准** : `go build ./...` 通过；配置文件加载返回合法 Config 结构体；环境变量可覆盖配置项
* **验证步骤** : `go build ./...`（V2）；启动服务验证配置加载无报错（V2）
* **测试要求** : L1 build 即可；V2 在 cc-apply 关闭为 `apply-covered`
* **依赖 / Wave** : wave-1，可与 Task 1 并行
* **回退方式** : 删除 `internal/config/` 和 `config.example.yaml`
* **完成后状态** : done
* **并发注意事项** : 与 Task 1 无文件冲突，可并行
* **Baseline / Delta** : `baseline/post-apply.json` — go build PASSED, go vet PASSED
* **对应 commit** : `2bd2ef1`

#### Task 3: SQLite Schema 初始化

* **目标** : 使用 `modernc.org/sqlite` 实现数据库连接与自动建表
* **不包含范围** : CRUD 操作封装（属于后续 change），不涉及 migration 版本管理
* **涉及文件** :
  * `internal/store/db.go` — 新建，DB 连接 + 建表
  * `internal/store/schema.go` — 新建，DDL 常量定义
* **关键签名** :
  ```go
  // internal/store/db.go
  // 新增
  func Open(path string) (*sql.DB, error)

  // internal/store/schema.go
  const Schema = `CREATE TABLE IF NOT EXISTS sessions (...)`
  ```
* **验收标准** : `go build ./...` 通过；启动后 `talkent.db` 文件生成；sessions/messages/analysis_reports 三表存在
* **验证步骤** : `go build ./...`（V3）；启动服务后 `sqlite3 talkent.db ".tables"` 确认三表（V3）
* **测试要求** : L1 build；V3 在 cc-apply 关闭为 `apply-covered`
* **依赖 / Wave** : wave-2，依赖 Task 2（需要 DBConfig）
* **回退方式** : 删除 `internal/store/` 和 `talkent.db`
* **完成后状态** : done
* **数据库注意事项** : 纯 Go SQLite 驱动（`modernc.org/sqlite`），无 CGO；建表使用 `IF NOT EXISTS` 保证幂等
* **Baseline / Delta** : `baseline/post-apply.json` — go build PASSED
* **对应 commit** : `2d103c9`

#### Task 4: HTTP Server + 健康检查

* **目标** : HTTP Server 启动、`/health` 端点、优雅关闭
* **不包含范围** : 业务 API 路由（属于后续各 change），不涉及中间件（auth/cors/logging 暂不实现）
* **涉及文件** :
  * `cmd/server/main.go` — 修改，组装依赖并启动 Server
  * `internal/server/server.go` — 新建，HTTP Server 初始化与路由注册
  * `internal/server/handlers.go` — 新建，`/health` handler
* **关键签名** :
  ```go
  // internal/server/server.go
  // 新增
  func New(cfg *config.Config, db *sql.DB, logger *slog.Logger) *http.Server

  // internal/server/handlers.go
  // 新增
  func healthHandler(w http.ResponseWriter, r *http.Request)
  ```
* **验收标准** : `go build ./...` 通过；服务启动监听配置端口；`curl /health` 返回 `{"status":"ok"}` 和 200
* **验证步骤** : `go build ./...`；启动服务 + `curl http://localhost:8080/health`（V4）
* **测试要求** : L1 build；V4 在 cc-apply 关闭为 `apply-covered`
* **依赖 / Wave** : wave-3，依赖 Task 2（Config）+ Task 3（DB）
* **回退方式** : 删除 `internal/server/`，`cmd/server/main.go` 回退到空 main
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-apply.json` — go build PASSED, E2E health check PASSED
* **对应 commit** : `fa02f4c`

#### Task 5: 结构化日志初始化

* **目标** : 统一 `log/slog` 初始化，配置时间格式（微秒）、级别、源码位置、输出目标
* **不包含范围** : 不在各模块加日志调用（仅 Server 启动和 health handler 加基础日志），不实现 request_id/trace_id
* **涉及文件** :
  * `internal/log/log.go` — 新建，Logger 初始化
  * `cmd/server/main.go` — 修改，调用 Logger 初始化
* **关键签名** :
  ```go
  // internal/log/log.go
  // 新增
  func New(cfg *config.LogConfig) *slog.Logger
  ```
* **验收标准** : `go build ./...` 通过；启动日志包含时间（微秒）、级别、源码位置；`/health` 请求有日志
* **验证步骤** : `go build ./...`（V5）；启动服务观察日志格式（V5）
* **测试要求** : L1 build；V5 在 cc-apply 关闭为 `apply-covered`
* **依赖 / Wave** : wave-3，依赖 Task 2（LogConfig）；建议与 Task 4 在同一 wave 执行
* **回退方式** : 删除 `internal/log/`
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-apply.json` — go build PASSED, log format verified
* **对应 commit** : `fa02f4c`

#### Task 6: 端到端集成验证

* **目标** : 组装所有模块，确认全链路启动正常，`main.go` 完成依赖注入
* **不包含范围** : 不新增功能代码，不编写自动化集成测试
* **涉及文件** :
  * `cmd/server/main.go` — 修改，完整组装流程
* **关键签名** :
  ```go
  // cmd/server/main.go
  func main() {
      // 1. 加载配置
      // 2. 初始化 Logger
      // 3. 打开数据库
      // 4. 创建 Server
      // 5. 启动 + 优雅关闭
  }
  ```
* **验收标准** : `go build ./...` 通过；服务启动无报错；`/health` 返回正常；数据库文件生成
* **验证步骤** : `go build ./...`；完整启动流程验证；汇总 V1-V6 全量通过（V1-V6）
* **测试要求** : L1 build；确认 V1-V6 全部 `apply-covered`
* **依赖 / Wave** : wave-4，依赖 Task 1-5 全部完成
* **回退方式** : 不涉及（组装 task，仅修改 main.go）
* **完成后状态** : done
* **Baseline / Delta** : `baseline/post-apply.json` — go build PASSED, E2E full-chain PASSED
* **对应 commit** : `fa02f4c`
