---
change_id: android-offline-cache
created: 2026-05-24
updated: 2026-05-24
---

### 任务拆分 — Android 离线缓存

#### 前置条件

* [ ] `android-core-shell` 已归档

#### Wave

```
Wave 1: T1 (Room entities + DAO + Database)
Wave 2: T2 (SessionRepo 缓存写入) + T3 (SessionRepo 缓存读取)
Wave 3: T4 (离线降级 UI) + T5 (build.gradle 依赖)
Wave 4: T6 (DAO 测试 + E2E)
```

#### Task 1: Room Entities + DAO + Database

* **目标**: 创建 Room schema（3 entity + 3 DAO + Database 类）
* **涉及文件**: `data/local/entity/SessionEntity.kt`, `MessageEntity.kt`, `ReportEntity.kt`, `data/local/dao/SessionDao.kt`, `MessageDao.kt`, `ReportDao.kt`, `data/local/TalkentDatabase.kt`
* **验收标准**: Room 编译通过，DAO 可执行 CRUD
* **验证步骤**: DAO 单元测试
* **测试要求**: DAO 测试（insert + query）
* **依赖 / Wave**: Wave 1
* **回退方式**: 删除 data/local/
* **完成后状态**: `done`
* **Baseline / Delta**: N/A

#### Task 2: SessionRepo 缓存写入

* **目标**: API 调用成功后写入 Room
* **涉及文件**: `data/repository/SessionRepo.kt`
* **验收标准**: 获取会话 → Room 中有数据；创建会话 → Room 中插入
* **验证步骤**: 集成测试
* **测试要求**: 无
* **依赖 / Wave**: Wave 2 (依赖 T1)
* **回退方式**: revert
* **完成后状态**: `done`
* **Baseline / Delta**: N/A

#### Task 3: SessionRepo 缓存读取 + 后台刷新

* **目标**: getMessages/getReport 先返回缓存，后台启动 API 刷新
* **涉及文件**: `data/repository/SessionRepo.kt`
* **验收标准**: 打开 Chat → 立即显示缓存消息 → 后台刷新
* **验证步骤**: 集成测试
* **测试要求**: 无
* **依赖 / Wave**: Wave 2
* **回退方式**: revert
* **完成后状态**: `done`
* **Baseline / Delta**: N/A

#### Task 4: 离线降级 UI

* **目标**: API 不可达时显示"离线模式"提示
* **涉及文件**: `ui/chat/ChatScreen.kt`, `ui/report/ReportScreen.kt`
* **验收标准**: 断网 → 显示缓存数据 + "离线模式" 提示
* **验证步骤**: 真机飞行模式测试
* **测试要求**: 无
* **依赖 / Wave**: Wave 3 (依赖 T3)
* **回退方式**: revert
* **完成后状态**: `done`
* **Baseline / Delta**: N/A

#### Task 5: Room 依赖 + DI

* **目标**: build.gradle 添加 Room 依赖 + TalkentApp 初始化 Database
* **涉及文件**: `app/build.gradle.kts`, `TalkentApp.kt`
* **验收标准**: 编译通过
* **验证步骤**: `./gradlew assembleDebug`
* **测试要求**: 无
* **依赖 / Wave**: Wave 3
* **回退方式**: revert
* **完成后状态**: `done`
* **Baseline / Delta**: N/A

#### Task 6: DAO 测试 + E2E

* **目标**: Room DAO 单元测试 + 真机离线验证
* **涉及文件**: `data/local/dao/*Test.kt`
* **验收标准**: DAO 测试通过 + 飞行模式下浏览历史
* **验证步骤**: 真机测试
* **测试要求**: DAO 单元测试
* **依赖 / Wave**: Wave 4 (依赖 T2-T5)
* **回退方式**: N/A
* **完成后状态**: `done`
* **Baseline / Delta**: N/A

#### Spec-Validation 覆盖

| 映射 | Task |
|------|------|
| V1 | T1, T6 |
| V2 | T2 |
| V3 | T3 |
| V4 | T4 |
