# analysis-engine — Log

## 2026-04-27 — cc-propose

### 决策记录

1. **分析触发方式**：手动 + 自动两者都支持
   - 手动：`POST /api/sessions/{id}/analyze`，用户主动请求
   - 自动：会话结束时通过 `OnSessionEndFunc` 回调触发，配置项控制开关
   - 原因：手动触发允许对话未结束也能预览分析；自动触发降低用户操作成本

2. **LLM 分析策略**：一次调用全维度
   - 原因：成本低、速度快，结构化 JSON Prompt 可保证输出格式
   - 容错：JSON 解析失败重试一次，重试 Prompt 强调格式要求

3. **重复分析**：允许，每次生成独立报告
   - 原因：同一会话可在不同时间点重新分析，观察 LLM 输出稳定性或 Prompt 调优效果
   - `analysis_reports` 表天然支持多记录

4. **报告格式**：JSON + Markdown
   - JSON 结构化数据供 API 消费和前端渲染
   - Markdown 全文供下载/阅读
   - 两者在同一报告记录中共存

5. **自动触发解耦设计**：`session.Service` 通过 `OnSessionEndFunc` 回调支持自动触发，不导入 `analysis` 包
   - `main.go` 负责将 `analysisSvc.TriggerAnalysis` 注入为回调
   - 自动触发失败不影响会话结束（Warn 日志）

6. **Schema 迁移**：新增 2 列（`markdown_content`、`model_used`），有默认值
   - 使用 `ALTER TABLE ADD COLUMN`，SQLite 3.35.0+ 支持
   - 列有默认值，回滚可删除列

7. **长对话 token 超限缓解**：分析 Prompt 中保留最近 N 轮完整对话 + 早期对话摘要（复用 memory.Manager 的 Summarize 能力）
   - MVP 不单独配置截断策略，复用已有的 memory_window_size 配置

### 澄清与假设

- 用户确认：手动 + 自动触发（非仅其中之一）
- 用户确认：一次调用全维度（非逐维度）
- 用户确认：允许重复分析
- 用户确认：JSON + Markdown 格式
- 假设：分析仅对 `completed` 会话生效，`active` 会话返回 409
- 假设：自动触发失败不阻断会话结束流程
- 假设：JSON 重试仅一次，不无限重试
- 假设：LLM 原始输出截断至 500 字符后记录到日志

### 触发的专题规则

- `rules/verification.md`：声明最低验证等级 L2，每个 task 映射验证 ID
- `rules/change-sizing.md`：分类为 M（4 个文件模块，一个主目标，清晰验证集群）
- `rules/database-changes.md`：涉及 Schema 迁移（新增 2 列），分类为 expand

### 自动校验结果

- 待运行 `cc-verify --harness-only --change analysis-engine`

## 2026-04-28 — cc-apply

### 实现记录

1. **T1: Analysis Store CRUD + Schema 迁移**
   - 新建 `internal/store/analysis.go`：`AnalysisReport` 结构体 + `AnalysisStore` CRUD（CreateReport/GetLatestReport/ListReports）
   - 修改 `internal/store/schema.go`：在 `CREATE TABLE analysis_reports` 中直接加入 `markdown_content` 和 `model_used` 列
   - 修改 `internal/store/db.go`：新增 `runMigrations()` 函数，对已存在数据库执行 `ALTER TABLE ADD COLUMN`，静默忽略"列已存在"错误

2. **T2: Analysis Engine**
   - 新建 `internal/analysis/engine.go`：`Engine` 结构体 + `Analyze` 方法
   - Prompt 构造：system prompt（JSON 格式要求）+ user prompt（角色+场景+维度+对话+格式提醒）
   - JSON 解析容错：`parseDimensionResults` 支持脱去 markdown code fence
   - 重试机制：`callWithRetry` 先调用一次，解析失败则追加助手+用户消息重试
   - Markdown 渲染：`renderMarkdown` 生成标题+会话信息+维度评分+综合评分

3. **T3: Analysis Service + 自动触发钩子**
   - 新建 `internal/analysis/service.go`：`Analyzer` 接口 + `Service` 结构体
   - `TriggerAnalysis`：校验会话状态→获取消息→解析配置→调用 Engine→持久化报告
   - 新建 `internal/analysis/errors.go`：`ErrSessionNotCompleted`、`ErrSessionNotFound`
   - 修改 `internal/session/service.go`：新增 `OnSessionEndFunc` 类型、`OnSessionEnd` 字段、`notifySessionEnd` 方法
   - `EndSession` 和 `Chat` 自动结束时都调用 `notifySessionEnd`

4. **T4: HTTP Handler + 配置 + main.go 接线 + 测试**
   - 新建 `internal/analysis/handler.go`：3 个端点（analyze/report/reports）
   - 修改 `internal/config/config.go`：新增 `AnalysisConfig`（`AutoTrigger bool`，默认 true）+ 环境变量 `TALKENT_ANALYSIS_AUTO_TRIGGER`
   - 修改 `cmd/server/main.go`：注入 analysis 依赖链 + `OnSessionEnd` 回调 + 路由注册
   - 新建 4 个测试文件：`analysis_test.go`（4 tests）、`engine_test.go`（7 tests）、`service_test.go`（6 tests）、`handler_test.go`（7 tests）

### 技术决策

- 使用 `Analyzer` 接口而非 `*Engine` 具体类型，使 Service 可测试
- Schema 迁移采用"新建表包含新列 + 对旧库执行 ALTER TABLE"双轨策略
- 自动触发回调使用 `context.Background()` 而非传播请求上下文，解耦请求生命周期
- 用户 Prompt 末尾追加"请严格按上述 JSON 格式输出分析结果"增强格式约束

### 验证证据

- `go build ./...` → Success
- `go test ./...` → 94 passed in 10 packages
- `go vet ./...` → No issues found
- V1-V12 全部 apply-covered
