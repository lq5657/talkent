---
id: android-mvvm-compose-template
type: technical_convention
status: candidate
applies_to:
  - cc-propose
  - cc-apply
triggers:
  - Android
  - Kotlin
  - Compose
  - MVVM
  - project structure
confidence: candidate
evidence:
  - .cc/changes/android-core-shell/
---

# Android MVVM + Compose 项目模板

## Rule / Insight

Android 客户端项目标准结构：Gradle Kotlin DSL + Compose BOM + MVVM (StateFlow) + Retrofit/OkHttp/Moshi + 手动 DI (Application 类) + EncryptedSharedPreferences。不使用 Hilt/Dagger 以减少初期复杂度。

## Project Structure

```
android/
├── build.gradle.kts          # AGP + Kotlin plugins
├── app/build.gradle.kts      # Compose BOM, Retrofit, Moshi, Navigation
└── app/src/main/java/<pkg>/
    ├── XxxApp.kt             # Application: manual DI container
    ├── MainActivity.kt       # Single Activity + setContent
    ├── data/
    │   ├── api/              # Retrofit interface + OkHttp interceptor + SSE client
    │   ├── model/            # Moshi DTOs (@JsonClass)
    │   └── repository/       # Repository (suspend fun + Result<T>)
    ├── ui/
    │   ├── theme/            # Material 3 Color + Theme
    │   ├── navigation/       # NavHost + routes
    │   ├── <feature>/        # Screen.kt + ViewModel.kt per feature
    │   └── ...
    └── util/                 # TokenManager, UrlConfig, etc.
```

## Applies When

- 新建 Android 客户端项目
- Kotlin + Compose 技术栈

## Does Not Apply When

- 已有 Hilt/Dagger DI 框架的项目
- 非 Compose 项目（XML Views）
- 多模块大型项目（需考虑 module 拆分）

## Evidence

- `android-core-shell` change: 30 个 Kotlin 文件，5 个页面

## Usage Notes

- 初期不需要 DI 框架；当 repository/screen 超过 10 个时再考虑引入 Hilt
- Compose BOM 统一管理版本，避免手动对齐
- Token 存储使用 EncryptedSharedPreferences（非 DataStore）以满足安全要求
