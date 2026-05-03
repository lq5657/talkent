---
alwaysApply: true
---
# Domain Language

This file records the shared domain vocabulary that future `cc-*` commands should use in specs, tasks, reviews, tests, and system explanations.

It is not split by programming language. Split it only by domain context or bounded context when a repo has multiple business vocabularies.

## Usage Rules

- Record domain concepts, product concepts, business states, and user-facing nouns that reduce ambiguity.
- Do not record general programming terms, framework names, package names, or language-specific implementation details.
- Prefer one canonical term. Put rejected aliases under `_Avoid_`.
- When a user uses a term that conflicts with this file, call out the conflict before freezing scope.
- If multiple contexts exist, keep this file as the root map and link to context-specific files such as `.cc/context/domain/ordering.md`.

## Terms

**角色 (Role)**: AI 对话者的人物设定，包括身份、性格、背景和语言风格，是对话训练中 AI 扮演的对象。在代码中为 `role.RoleType` 枚举。
_Avoid_: AI角色, 对话角色, AI对话者

**场景 (Scenario)**: 对话发生的具体情境，包括地点、时间、关系和对话目的，由角色设定自然衍生。与 RoleDescription 一起存储在 session.role_config JSON 中。

**角色类型 (RoleType)**: 角色的分类标签，如"技术面试官"、"聊天伙伴"、"数学老师"等，用于匹配训练目标和分析维度模板。
_Avoid_: 角色类别

**训练目标 (Training Goal)**: 用户通过本次对话希望提升的表达能力维度，如说服力、共情、逻辑清晰度等。由 `POST /api/roles/recommend-goals` 根据角色描述通过 LLM 推荐。存储为 `role.TrainingGoal` (Name + Description)。
_Avoid_: 学习目标, 练习目标

**分析维度 (Analysis Dimension)**: 评估对话表达能力的结构化维度，如逻辑性、说服力、情绪控制、语言流畅度等。由 `POST /api/roles/recommend-dimensions` 推荐。存储为 `role.Dimension` (Name + Description)，分析结果存储为 `dimension_results` JSON。
_Avoid_: 评估维度, 评价维度

**会话 (Session)**: 一次完整的 1v1 对话训练单元，包含角色设定、训练目标、分析维度和多轮对话。有明确的生命周期状态 (active → completed)。存储于 `sessions` 表。
_Avoid_: 对话, 聊天, conversation

**会话状态 (Session Status)**: 会话的生命周期阶段。当前仅两个状态: `active` (进行中) 和 `completed` (已结束)。没有独立的 "analyzed" 状态。状态变更通过 `store.UpdateSessionStatus()` 集中执行。
_Avoid_: 会话阶段

**对话轮数 (Round)**: 会话中的一问一答单元（user 消息 + assistant 回复 = 1 轮）。`currentRound = msgCount / 2`。会话有 `round_limit` 上限，达到后自动结束。Chat API 返回 `round_info` (current/limit/is_last)。
_Avoid_: 轮次, 消息数

**记忆窗口 (Memory Window)**: 滑动窗口机制 (`memory.Manager`)，默认保留最近 10 条消息。超出窗口的历史对话触发 LLM 摘要压缩。
_Avoid_: 上下文窗口, 历史消息

**记忆来源 (MemorySource)**: 表示当前 LLM 请求使用的上下文构建方式。`"window"` 表示全部消息在窗口内；`"summary+window"` 表示窗口外消息已压缩为摘要。
_Avoid_: 上下文模式

**分析报告 (Analysis Report)**: 会话结束后由分析引擎生成的 Markdown 格式多维度分析反馈，包含各维度的评分、具体例子和改进建议。存储于 `analysis_reports` 表，通过 `GET /api/sessions/:id/report` 获取。
_Avoid_: 评估报告, 反馈报告

**分析引擎 (Analysis Engine)**: 负责调用 LLM 执行多维度分析、生成结构化 Markdown 报告的核心组件 (`analysis.Engine`)。
_Avoid_: 评估引擎

**自动触发分析 (Auto Trigger)**: 会话结束后由 `OnSessionEnd` callback 自动启动分析（使用 `context.Background()`），由 `analysis.auto_trigger` 配置控制。
_Avoid_: 自动分析

**摘要压缩 (Summary Compression)**: 当对话消息数超过记忆窗口大小时，调用 LLM (temperature=0.3) 将早期消息压缩为摘要文本，附加到 system prompt 中。压缩失败时降级为仅使用窗口内消息。

**系统提示词 (System Prompt)**: 由 `session.BuildSystemPrompt()` 构建，包含角色描述、场景、训练目标和分析维度，作为 LLM 对话的 system message。

## Relationships

- **角色 (Role)** 定义 → 推荐 **训练目标 (Training Goal)** → 推荐 **分析维度 (Analysis Dimension)**
- **会话 (Session)** 包含 角色配置 + 训练目标 + 分析维度 + 多轮对话 + 分析报告
- **记忆窗口 (Memory Window)** 管理会话中的对话上下文，超出窗口触发 **摘要压缩 (Summary Compression)**
- **分析引擎 (Analysis Engine)** 基于会话完整内容生成 **分析报告 (Analysis Report)**
- **会话状态 (Session Status)** active → completed (手动 EndSession 或自动 RoundLimit 触发)
- **自动触发分析 (Auto Trigger)** 依赖 `OnSessionEnd` callback，由 completed 状态变更触发

## Flagged Ambiguities

- 无待确认的歧义术语。

## Context Splits

Use this section only when one glossary would mix unrelated domain languages.

| Context | Glossary | Scope | Relationship |
|---------|----------|-------|--------------|
| default | this file | Whole project or single domain context | none |
