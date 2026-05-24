---
change_id: android-core-shell
created: 2026-05-24
updated: 2026-05-24
---

### 任务拆分 — Android 客户端骨架

#### 前置条件

* [ ] `spec.md` 已确认且 `status = propose`
* [ ] `android-backend-ready` 已归档（done）

#### 依赖 / Wave 总览

```
Wave 1: T1 (Gradle 项目骨架)
Wave 2: T2 (Data models) + T3 (API + Auth repo)
Wave 3: T4 (Theme + Navigation)
Wave 4: T5 (Setup) + T6 (Chat) + T7 (Report)
Wave 5: T8 (Settings) + T9 (Token refresh)
Wave 6: T10 (E2E verification)
```

#### 文件变更清单

| 文件 | 操作 | 涉及 Task |
|------|------|-----------|
| android/ (Gradle 完整项目) | 新增 | T1 |
| app/src/.../data/model/*.kt | 新增 | T2 |
| app/src/.../data/api/TalkentApi.kt | 新增 | T3 |
| app/src/.../data/api/AuthInterceptor.kt | 新增 | T3 |
| app/src/.../data/api/SseClient.kt | 新增 | T3 |
| app/src/.../data/repository/AuthRepo.kt | 新增 | T3 |
| app/src/.../data/repository/SessionRepo.kt | 新增 | T3 |
| app/src/.../util/TokenManager.kt | 新增 | T3 |
| app/src/.../ui/theme/*.kt | 新增 | T4 |
| app/src/.../ui/navigation/NavGraph.kt | 新增 | T4 |
| app/src/.../MainActivity.kt | 新增 | T4 |
| app/src/.../ui/setup/*.kt | 新增 | T5 |
| app/src/.../ui/chat/*.kt | 新增 | T6 |
| app/src/.../ui/report/*.kt | 新增 | T7 |
| app/src/.../ui/settings/*.kt | 新增 | T8 |
| app/src/.../util/UrlConfig.kt | 新增 | T8 |

#### Task 1: Gradle 项目骨架

* **目标**: 创建可编译的空 Android 项目
* **涉及文件**: `android/build.gradle.kts`, `settings.gradle.kts`, `gradle.properties`, `app/build.gradle.kts`, `AndroidManifest.xml`
* **验收标准**: `./gradlew assembleDebug` 成功
* **验证步骤**: `cd android && ./gradlew assembleDebug`
* **测试要求**: 无
* **依赖 / Wave**: Wave 1
* **回退方式**: 删除 android/ 目录
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A

#### Task 2: Data Models

* **目标**: 定义 API DTO（Kotlin data class + Moshi）
* **涉及文件**: `app/src/.../data/model/Models.kt`
* **验收标准**: 所有 DTO 可序列化/反序列化
* **验证步骤**: `./gradlew test` (JVM Moshi adapter tests)
* **测试要求**: Moshi adapter 测试
* **依赖 / Wave**: Wave 2
* **回退方式**: 删除 model 文件
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A

#### Task 3: API Client + Repository + TokenManager

* **目标**: Retrofit 接口 + OkHttp interceptor + SSE client + AuthRepo + SessionRepo + TokenManager
* **涉及文件**: `data/api/TalkentApi.kt`, `AuthInterceptor.kt`, `SseClient.kt`, `data/repository/AuthRepo.kt`, `SessionRepo.kt`, `util/TokenManager.kt`
* **验收标准**: login 成功 → token 存储 → API 携带 Bearer → 401 自动 refresh
* **验证步骤**: `./gradlew test` + 连接后端手工验证 login 流程
* **测试要求**: Repository + Interceptor 单元测试
* **依赖 / Wave**: Wave 2
* **回退方式**: 删除 data/ 和 util/ 目录
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A

#### Task 4: Theme + Navigation + MainActivity

* **目标**: Material 3 主题 + NavHost (4 routes) + 单 Activity
* **涉及文件**: `ui/theme/Theme.kt`, `ui/theme/Color.kt`, `ui/navigation/NavGraph.kt`, `MainActivity.kt`
* **验收标准**: App 启动 → 4 个路由可跳转
* **验证步骤**: 编译安装到模拟器 → 导航测试
* **测试要求**: 无
* **依赖 / Wave**: Wave 3
* **回退方式**: 删除 ui/ 目录
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A

#### Task 5: Setup Screen

* **目标**: 角色设定页 (描述输入 → 推荐目标 → 推荐维度 → 创建会话)
* **涉及文件**: `ui/setup/SetupScreen.kt`, `ui/setup/SetupViewModel.kt`
* **验收标准**: 完整流程 → 创建会话 → 跳转 chat/{id}
* **验证步骤**: ViewModel 单元测试 + 手工验证完整流程
* **测试要求**: ViewModel 单元测试 (mock repo)
* **依赖 / Wave**: Wave 4
* **回退方式**: 删除 ui/setup/
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A

#### Task 6: Chat Screen

* **目标**: 对话页 (消息列表 + SSE 流式 + 轮数 + 结束)
* **涉及文件**: `ui/chat/ChatScreen.kt`, `ui/chat/ChatViewModel.kt`
* **验收标准**: 发送消息 → 流式显示 → 多轮 → 结束 → 跳转 report/{id}
* **验证步骤**: ViewModel 单元测试 + 手工验证 SSE 流式
* **测试要求**: ViewModel 单元测试
* **依赖 / Wave**: Wave 4
* **回退方式**: 删除 ui/chat/
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A

#### Task 7: Report Screen

* **目标**: 报告页 (维度卡片 + 报告文本)
* **涉及文件**: `ui/report/ReportScreen.kt`, `ui/report/ReportViewModel.kt`
* **验收标准**: 加载报告 → 显示维度卡片 + 文本
* **验证步骤**: ViewModel 单元测试 + 手工验证报告展示
* **测试要求**: ViewModel 单元测试
* **依赖 / Wave**: Wave 4
* **回退方式**: 删除 ui/report/
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A

#### Task 8: Settings Screen + URL Config

* **目标**: 设置页 (后端地址 + 测试连接 + 登出)
* **涉及文件**: `ui/settings/SettingsScreen.kt`, `ui/settings/SettingsViewModel.kt`, `util/UrlConfig.kt`
* **验收标准**: 修改地址 → 测试连接 → 登出 → 回到登录
* **验证步骤**: 手工验证设置页
* **测试要求**: 无
* **依赖 / Wave**: Wave 5
* **回退方式**: 删除 ui/settings/
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A

#### Task 9: Token Auto-Refresh + Login Flow

* **目标**: 全局 token 过期检测 + 自动 refresh + 登录页 + 路由守卫
* **涉及文件**: `util/TokenManager.kt`(完善), `ui/login/LoginScreen.kt`, `ui/navigation/NavGraph.kt`(更新)
* **验收标准**: 401 → refresh → 重试成功；refresh 过期 → 登录页；无 token → 登录页
* **验证步骤**: 单元测试 + 手工验证 token 生命周期
* **测试要求**: TokenManager + Interceptor 单元测试
* **依赖 / Wave**: Wave 5
* **回退方式**: 移除 TokenManager
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A

#### Task 10: E2E Manual Verification

* **目标**: APK 全流程手工验证 + 截图记录
* **涉及文件**: 无（验证 task）
* **验收标准**: 6 步流程全部走通 (登录 → Setup → Chat → Report → 设置 → 登出)
* **验证步骤**: 安装 APK 到模拟器/真机 → 按步骤执行 → 截图记录
* **测试要求**: 截图 + 文字记录
* **依赖 / Wave**: Wave 6
* **回退方式**: N/A
* **完成后状态**: `done`
* **Baseline / Delta（按需）**: N/A

#### Spec-Validation 覆盖映射

| Spec 映射编号 | 覆盖 Task | 说明 |
|---------------|-----------|------|
| V1 | T3, T9 | 登录 + token 管理 |
| V2 | T5 | 角色设定流程 |
| V3 | T6 | SSE 流式对话 |
| V4 | T7 | 分析报告展示 |
| V5 | T8 | 设置页面 |
| V6 | T3, T9 | Token 自动 refresh |
