---
change_id: android-voice
created: 2026-05-24
updated: 2026-05-24
---

### 任务拆分 — Android 语音交互

#### 前置条件

* [ ] `spec.md` 已确认
* [ ] `android-core-shell` 已归档

#### 依赖 / Wave

```
Wave 1: T1 (SpeechRecorder util) + T2 (TtsPlayer util)
    │
Wave 2: T3 (ChatScreen UI modification) + T4 (ViewModel + Manifest)
    │
Wave 3: T5 (E2E manual verification on real device)
```

#### 文件变更清单

| 文件 | 操作 | 涉及 Task |
|------|------|-----------|
| util/SpeechRecorder.kt | 新增 | T1 |
| util/TtsPlayer.kt | 新增 | T2 |
| ui/chat/ChatScreen.kt | 修改 | T3 |
| ui/chat/ChatViewModel.kt | 修改 | T4 |
| AndroidManifest.xml | 修改 | T4 |

#### Task 1: SpeechRecorder

* **目标**: 封装 SpeechRecognizer + 录音状态管理
* **涉及文件**: `util/SpeechRecorder.kt`
* **关键设计**: 封装 `SpeechRecognizer.createSpeechRecognizer()` + `RecognizerIntent`
* **验收标准**: 调用 `startListening()` → 识别回调 → `onResult(text)` → `stopListening()`
* **验证步骤**: 真机测试
* **测试要求**: 无（依赖硬件）
* **依赖 / Wave**: Wave 1
* **回退方式**: 删除文件
* **完成后状态**: `done`
* **Baseline / Delta**: N/A

#### Task 2: TtsPlayer

* **目标**: 封装 TextToSpeech + 音频焦点管理
* **涉及文件**: `util/TtsPlayer.kt`
* **关键设计**: 封装 `TextToSpeech` + `AudioManager` 焦点申请
* **验收标准**: `speak(text)` → TTS 朗读 → `stop()` 停止
* **验证步骤**: 真机测试
* **测试要求**: 无（依赖硬件）
* **依赖 / Wave**: Wave 1
* **回退方式**: 删除文件
* **完成后状态**: `done`
* **Baseline / Delta**: N/A

#### Task 3: ChatScreen UI 修改

* **目标**: 添加录音按钮 + 按住手势 + 录音动画
* **涉及文件**: `ui/chat/ChatScreen.kt`
* **UI 改动**: 输入框旁加 🎤 按钮，`pointerInput` 检测长按/释放
* **验收标准**: 按住显示录音中动画 → 松开发送 → 禁用发送按钮
* **验证步骤**: 真机 UI 测试
* **测试要求**: 无
* **依赖 / Wave**: Wave 2 (依赖 T1+T2)
* **回退方式**: revert ChatScreen.kt
* **完成后状态**: `done`
* **Baseline / Delta**: N/A

#### Task 4: ChatViewModel + Manifest

* **目标**: 语音状态管理 + 权限声明 + TTS 自动朗读触发
* **涉及文件**: `ui/chat/ChatViewModel.kt`, `AndroidManifest.xml`
* **验收标准**: 识别结果自动填入 → 发送 → SSE 完成 → TTS 自动朗读
* **验证步骤**: 真机全流程测试
* **测试要求**: 无
* **依赖 / Wave**: Wave 2 (依赖 T1+T2)
* **回退方式**: revert
* **完成后状态**: `done`
* **Baseline / Delta**: N/A

#### Task 5: E2E 真机验证

* **目标**: APK 安装到真机，全流程录屏验证
* **涉及文件**: 无（验证 task）
* **验收标准**: 全流程走通 + 录屏证据
* **验证步骤**: 按住录音 → 识别发送 → SSE 流式回复 → TTS 朗读 → 打断测试 → 权限拒绝测试
* **测试要求**: 录屏 + 截图
* **依赖 / Wave**: Wave 3 (依赖 T3+T4)
* **回退方式**: N/A
* **完成后状态**: `done`
* **Baseline / Delta**: N/A

#### Spec-Validation 覆盖映射

| 映射 | 覆盖 Task |
|------|-----------|
| V1 | T1, T3, T4 |
| V2 | T2, T3, T4 |
| V3 | T2, T3 |
| V4 | T4 |
