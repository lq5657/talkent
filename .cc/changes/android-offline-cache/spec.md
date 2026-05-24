---
change_id: android-offline-cache
status: propose
depends_on: [android-core-shell]
parallel_safe: false
branch: feat/android-offline-cache
created: 2026-05-24
updated: 2026-05-24
complexity: medium
proposal_profile: standard
verification_level: L3
evidence_types: [package, chain, manual]
---

### Android 离线缓存 — Room 数据库

#### 0.1 Proposal Profile

| Profile | 条件 | 本次 |
|---------|------|------|
| standard | 单一 scope，Room 集成 | ✓ |

#### 1. 背景与目标

`android-core-shell` 已实现完整文字对话。本 change 使用 Room 数据库缓存会话、消息和报告，实现"先显示缓存，后台刷新"模式，离线时可浏览历史数据。

#### 1.1 本次不做

- 离线消息发送队列（离线时仍需联网对话）
- 缓存过期/清理策略（MVP 阶段使用全量缓存）

#### 2. 代码现状

当前所有数据直接从 API 获取。需要新增 Room database + DAO + entity 层，并修改 repository 支持缓存优先策略。

#### 3. 功能点

* [ ] F1: Room 数据库 — SessionEntity, MessageEntity, ReportEntity + DAO + Database
* [ ] F2: 缓存写入 — API 成功返回后写入 Room
* [ ] F3: 缓存读取 — 打开 Chat/Report 页面时优先显示缓存数据
* [ ] F4: 后台刷新 — 缓存显示后异步从 API 拉取最新数据更新

#### 4. 业务规则

- 缓存为全量追加（不做过期清理）
- active 会话的消息也缓存（重新打开时快速恢复）
- API 失败时显示缓存数据 + 提示"离线模式"

#### 5. 数据变更

* **是否涉及 migration**：否（新数据库，首次创建）
* **Room schema**：3 张表

#### 6. 接口变更

* **是否涉及对外契约变更**：否

#### 7. 影响范围

| 文件 | 操作 | 说明 |
|------|------|------|
| `data/local/entity/*.kt` | 新增 | Room Entity |
| `data/local/dao/*.kt` | 新增 | Room DAO |
| `data/local/Database.kt` | 新增 | Room Database |
| `data/repository/SessionRepo.kt` | 修改 | 缓存写入+读取 |
| `app/build.gradle.kts` | 修改 | Room 依赖 |
| `TalkentApp.kt` | 修改 | 初始化 Database |

#### 8. 风险

| 类型 | 描述 | 处理 |
|------|------|------|
| 性能 | 大量消息可能影响 Room 查询 | 按 session_id 索引分页 |
| 一致性 | 缓存可能落后于服务端 | 先显示缓存后刷新 |

#### 9. 测试策略

* **最低验证等级**：L3（Room DAO 测试 + 缓存链路验证）
* **验证证据**：Room DAO 单元测试 + 真机离线模式验证

#### 9.1 需求-验证映射

| 编号 | 需求项 | 等级 | 证据类型 | Task | 状态 |
|------|--------|------|----------|------|------|
| V1 | Room database + DAO | L2 | package | Room DAO 测试 | T1 | todo |
| V2 | 缓存写入（API→Room） | L3 | chain | API→Room 写入链路 | T2 | todo |
| V3 | 缓存优先读取 + 后台刷新 | L3 | chain | 缓存→后台刷新链路 | T3 | todo |
| V4 | 离线降级提示 | L4 | manual | 飞行模式验证 | T4 | todo |

#### 12. 技术决策

| 决策 | 选择 | 放弃 | 原因 |
|------|------|------|------|
| 数据库 | Room | SQLDelight | Google 官方，Compose 生态 |
| 缓存策略 | 缓存优先 + 后台刷新 | 仅离线降级 | 用户选择 |
| Migration | 首次创建，无 migration | — | MVP |

#### 15. 确认记录（HARD-GATE）

* **confirmed_at**：2026-05-24
* **confirmed_by**：user
* **confirmed_spec_revision**：2026-05-24
* **confirmed_tasks_revision**：2026-05-24
* **confirmed_scope**：Room 缓存 — 会话+消息+报告，6 task
* **resolved_risk_decisions**：Room 首次创建数据库，缓存优先+后台刷新
* **accepted_residual_risks**：无过期清理策略，全量缓存
* **human_review_required**：true
* **human_review_status**：approved
