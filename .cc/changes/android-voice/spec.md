---
change_id: android-voice
status: review
depends_on: [android-core-shell]
parallel_safe: false
branch: feat/android-voice
created: 2026-05-24
updated: 2026-05-24
complexity: medium
proposal_profile: standard
verification_level: L4
evidence_types: [manual]
---

### Android 语音交互 — STT + TTS

#### 0.1 Proposal Profile

| Profile | 使用条件 | 本次 |
|---------|----------|------|
| standard | 多阶段确认，单一 scope | ✓ |

#### 1. 背景与目标

`android-core-shell` 已实现文字对话。本 change 添加语音输入（STT）和语音合成（TTS），使用 Android 系统内置 API（SpeechRecognizer + TextToSpeech），零额外依赖。

#### 1.1 本次不做

- 连续对话模式（保持 push-to-talk）
- 自定义语音识别模型
- 服务端 TTS
- 离线语音识别（依赖设备是否支持）

#### 2. 代码现状

`ChatScreen.kt` 和 `ChatViewModel.kt` 已实现文字对话 + SSE 流式接收。本次在其基础上添加语音交互层。

#### 3. 功能点

* [ ] F1: 语音输入 — 按住录音按钮 → SpeechRecognizer 录音 → 松手填入输入框 → 自动发送
* [ ] F2: 语音合成 — 收到 assistant 完整回复后 → TTS 自动朗读
* [ ] F3: 音频焦点 — 朗读中按住录音 → 停止 TTS → 开始新录音（互斥）
* [ ] F4: UI 状态 — 录音时显示波纹动画/音量指示 + 禁用发送按钮

#### 4. 业务规则

- 录音权限：首次使用时请求 `RECORD_AUDIO`，拒绝时降级为纯文字模式
- TTS 初始化：异步检查中文语音包可用性，不可用时静默跳过
- 音频焦点：录音优先于朗读（用户输入打断 AI 输出）
- 错误处理：识别失败时在输入框中保留用户已识别部分文字（如有）

#### 5. 数据变更

* **是否涉及 migration**：否

#### 6. 接口变更

* **是否涉及对外契约变更**：否（纯客户端变更）

#### 7. 影响范围

| 文件 | 操作 | 说明 |
|------|------|------|
| `app/src/.../util/SpeechRecorder.kt` | 新增 | SpeechRecognizer 封装 |
| `app/src/.../util/TtsPlayer.kt` | 新增 | TextToSpeech 封装 |
| `app/src/.../ui/chat/ChatScreen.kt` | 修改 | 录音按钮 + 手势 + 动画 |
| `app/src/.../ui/chat/ChatViewModel.kt` | 修改 | 语音状态管理 |
| `app/src/.../AndroidManifest.xml` | 修改 | RECORD_AUDIO 权限 |

#### 8. 风险

| 类型 | 描述 | 处理方式 |
|------|------|----------|
| 权限 | 用户拒绝录音权限 | 降级为纯文字模式，显示提示 |
| 兼容 | 部分设备 TTS 不支持中文 | 检查语言包，不可用时静默跳过 |
| 兼容 | SpeechRecognizer 在模拟器不可用 | 真机测试，模拟器使用文字输入 |

#### 9. 测试策略

* **最低验证等级**：L4（manual — 需真机验证语音 I/O）
* **验证证据要求**：真机录屏 + APK 安装验证

#### 9.1 需求-验证映射

| 编号 | 需求项 | 等级 | 证据类型 | 建议验证动作 | Task | 状态 |
|------|--------|------|----------|-------------|------|------|
| V1 | 语音输入（STT） | L4 | manual | 按住录音 → 识别 → 发送 → SSE 回复 | T2 | todo |
| V2 | 语音合成（TTS） | L4 | manual | 收到回复 → 自动朗读 | T3 | todo |
| V3 | 音频焦点互斥 | L4 | manual | TTS 播放中按住录音 → TTS 停止 | T4 | todo |
| V4 | 权限降级 | L4 | manual | 拒绝录音权限 → 纯文字模式不崩溃 | T4 | todo |

#### 12. 技术决策

| 决策 | 选择 | 放弃 | 原因 |
|------|------|------|------|
| STT | Android SpeechRecognizer | 第三方 SDK | 零依赖，系统内置 |
| TTS | Android TextToSpeech | 服务端 TTS | 免费，低延迟 |
| 交互 | Push-to-talk | 连续模式 | 用户选择 |

#### 15. 确认记录（HARD-GATE）

* **confirmed_at**：2026-05-24
* **confirmed_by**：user
* **confirmed_spec_revision**：2026-05-24
* **confirmed_tasks_revision**：2026-05-24
* **confirmed_scope**：Android 语音交互 (STT/TTS)，5 task，3 wave，5 文件
* **resolved_risk_decisions**：SpeechRecognizer + TextToSpeech + Push-to-talk
* **accepted_residual_risks**：模拟器不可用语音（真机验证），TTS 中文降级
* **human_review_required**：true
* **human_review_status**：approved
