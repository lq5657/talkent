# Analysis Engine 模式

> 来源: `analysis-engine` change 归档

## OnSessionEndFunc 回调解耦

- `session.Service` 暴露 `OnSessionEnd OnSessionEndFunc` 字段，由 `main.go` 注入具体回调
- `session` 包不导入 `analysis` 包，避免循环依赖
- 回调使用 `context.Background()` 解耦请求生命周期，自动触发失败只记 Warn 不阻断
- **Chat 自动结束和 EndSession 手动结束都必须调用 `notifySessionEnd`**，遗漏任一路径会导致自动触发失效
- 测试模式：直接设置 `svc.OnSessionEnd` 字段，验证回调被调用及参数正确

## SQLite Migration 双轨策略

- `CREATE TABLE` 语句包含所有列（含新增列），保证新数据库一次建表完整
- `runMigrations()` 对已存在数据库执行 `ALTER TABLE ADD COLUMN`
- 用 `isDuplicateColumnError()` 函数区分"列已存在"（可忽略）与其他 ALTER TABLE 错误（必须上抛）
- 新增列必须有 `NOT NULL DEFAULT`，保证回滚时删除列安全
- 此模式适用于 SQLite expand 类型迁移；contract/migrate 类型需更复杂的兼容窗口策略

## 报告双重格式

- `analysis_reports` 表同时存储 `dimension_results`（JSON）和 `markdown_content`（Markdown 全文）
- JSON 供 API 程序化消费和前端渲染；Markdown 供人类阅读和下载
- Handler 的 `reportToResponse` 需处理 `dimension_results` 反序列化错误，设置 `dims = nil` 而非静默吞错
