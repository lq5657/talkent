### 变更日志 — Android 语音交互

#### 时间线

| 时间 | 阶段 | 事件 |
|------|------|------|
| 2026-05-24 | propose | cc-propose 创建 change — 阶段 3: 语音交互 |

#### 技术决策

| 决策 | 选择 | 放弃 | 原因 |
|------|------|------|------|
| STT | Android SpeechRecognizer | 第三方 SDK | 零依赖 |
| TTS | Android TextToSpeech | 服务端 TTS | 免费低延迟 |
| 交互 | Push-to-talk | 连续模式 | 用户选择 |
