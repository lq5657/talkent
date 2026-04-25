### 变更日志 — 项目脚手架搭建

#### 时间线

| 时间 | 阶段 | 事件 | 备注 |
|------|------|------|------|
| 2026-04-25 | propose | `cc-propose scaffold-project` 创建提案 | |
| 2026-04-25 | apply | `cc-apply` 6 tasks 全部完成，4 个 commit | go build/test/vet + E2E 通过 |
| 2026-04-25 | review | `cc-review` Stage 1+2 PASSED，0 findings，可归档 | |
| 2026-04-25 | done | `cc-archive` 归档完成，knowledge 无需沉淀 | 待合并到主分支 |

#### 技术决策

| 决策 | 选择 | 放弃的方案 | 原因 |
|------|------|-----------|------|
| HTTP 路由 | `net/http` 标准库 | chi, gin | MVP 接口少，零依赖优先 |
| SQLite 驱动 | `modernc.org/sqlite` | `mattn/go-sqlite3` | 纯 Go 无 CGO，跨平台编译简单 |
| 配置格式 | YAML | JSON, TOML | 可读性好，Go 生态主流 |
| 配置加载 | 文件 + 环境变量覆盖 | 仅文件/仅环境变量 | 兼顾便捷与安全 |
| 日志库 | `log/slog` 标准库 | zap, zerolog | 零依赖，满足 MVP 需求 |
