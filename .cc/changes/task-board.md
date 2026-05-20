# Task Board

本文件记录 change 级工作看板，用于快速判断当前仓库有哪些候选、进行中、阻塞或待归档的工作。
它不是 `spec.md` / `tasks.md` 的替代品；真实需求、验收和证据仍以单个 `.cc/changes/<change-id>/` 为准。

```text
last_updated: 2026-05-20
updated_by: cc-propose
```

## 1. 正式 Change

| change_id | 状态 | 来源 | 目标摘要 | 影响模块 | 阻塞 / 依赖 | 下一命令 | 最近证据 |
|-----------|------|------|----------|----------|-------------|----------|----------|
| message-timing | done | cc-archive | 对话页消息显示开始时间、结束时间、持续时间 | session handler/service, ChatView, MessageBubble | 无 | — | 已归档，0 open findings |
| voice-interaction | done | cc-archive | 流式文字 + 浏览器语音交互（STT/TTS） | llm, session, server, ChatView, ChatInput, MessageBubble | 无 | — | 5 findings fixed, 3 knowledge entries, L4 verified |

## 2. Backlog 候选

| 候选项 | 来源 | 推荐 change_id | 价值 | 前置条件 | 建议下一步 |
|--------|------|----------------|------|----------|------------|
| 待填充 | 待填充 | 待填充 | 待填充 | 待填充 | cc-propose |

## 3. 阻塞项

| change_id / 候选项 | 阻塞原因 | 需要谁确认 | 恢复入口 | 记录位置 |
|--------------------|----------|------------|----------|----------|
| 待填充 | 待填充 | 待填充 | 待填充 | 待填充 |

## 更新规则

- `cc-new-project` 可写入 Backlog 候选，但不得创建正式 change。
- `cc-propose` 创建正式 change 后，必须新增或更新正式 Change 行。
- `cc-apply`、`cc-test`、`cc-review`、`cc-fix` 和 `cc-archive` 必须同步状态、阻塞项和下一命令。
- 看板只保存摘要和导航，不复制 spec/tasks/review 的完整正文。
