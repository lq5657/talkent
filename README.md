# Talkent

角色扮演对话训练智能体。自由设定角色和场景，与 AI 进行 1v1 沉浸式对话训练，结束后获得多维度、结构化的表达分析反馈。

## 技术栈

| 层 | 技术 |
|---|------|
| 后端 | Go 1.24+, net/http, log/slog |
| Web 前端 | Vue 3 (Composition API), TypeScript, Tailwind CSS v4, Vite |
| Android | Kotlin, Jetpack Compose, Retrofit, Room |
| 数据库 | SQLite (后端) + Room (Android 离线缓存) |
| LLM | OpenAI 兼容 API（DeepSeek 等） |
| 认证 | JWT (access + refresh token) |

## 快速开始

**前置要求**：Go 1.24+, Node.js 20+, Android SDK 35+（可选，仅 Android 客户端）

```bash
# 1. 配置
cp config.example.yaml config.yaml
# 编辑 config.yaml，填入 llm.api_key，或通过环境变量注入：
# export TALKENT_LLM_API_KEY="your-api-key"
# export TALKENT_AUTH_JWT_SECRET="your-secret-key"  # 生产环境必须修改

# 2. 启动后端
make run

# 3. 构建前端（可选，后端已托管静态文件）
cd web && npm install && npm run build && cd ..

# 4. 浏览器访问 http://localhost:8080
# 默认登录凭据: admin / admin
```

### Android 客户端

**前置要求**：Android SDK 35+, JDK 17+, Gradle 8.9+（或使用项目自带 `gradlew`）

```bash
# 1. 设置 Android SDK 路径
export ANDROID_HOME=~/android-sdk

# 2. 编译 Debug APK（18.7MB，开发调试用）
cd android && ./gradlew assembleDebug
# 产物: android/app/build/outputs/apk/debug/app-debug.apk

# 3. 编译 Release APK（需签名，生产发布用）
./gradlew assembleRelease
# 产物: android/app/build/outputs/apk/release/app-release-unsigned.apk
# 正式发布需配置 signingConfig 签名

# 4. 运行单元测试（32 tests）
./gradlew test

# 5. 安装到模拟器或真机
adb install app/build/outputs/apk/debug/app-debug.apk
```

默认连接 `http://10.0.2.2:8080`（Android 模拟器到宿主机的约定映射），真机请在 App 设置页面配置后端 IP。

## 配置

```yaml
# config.yaml
server:
  host: "0.0.0.0"
  port: 8080
  request_timeout: 30s

database:
  path: "./talkent.db"

llm:
  provider: "deepseek"
  base_url: "https://api.deepseek.com/v1"
  api_key: ""               # 建议通过 TALKENT_LLM_API_KEY 环境变量注入
  model: "deepseek-chat"
  timeout: 30s

auth:
  username: "admin"         # 建议通过 TALKENT_AUTH_USERNAME 环境变量注入
  password: "admin"         # 建议通过 TALKENT_AUTH_PASSWORD 环境变量注入
  jwt_secret: "change-me-in-production"  # 生产环境必须通过 TALKENT_AUTH_JWT_SECRET 覆盖
  access_expiry: 1h
  refresh_expiry: 168h      # 7 天

log:
  level: "info"
  file: ""                  # 空字符串 = stdout

session:
  memory_window_size: 10    # 滑动窗口消息数

analysis:
  auto_trigger: true        # 会话结束后自动触发分析
```

支持的环境变量覆盖（优先级高于 config.yaml）：

| 环境变量 | 对应配置 |
|----------|----------|
| `TALKENT_LLM_API_KEY` | `llm.api_key` |
| `TALKENT_LLM_BASE_URL` | `llm.base_url` |
| `TALKENT_LLM_MODEL` | `llm.model` |
| `TALKENT_SERVER_PORT` | `server.port` |
| `TALKENT_AUTH_USERNAME` | `auth.username` |
| `TALKENT_AUTH_PASSWORD` | `auth.password` |
| `TALKENT_AUTH_JWT_SECRET` | `auth.jwt_secret` |
| `TALKENT_AUTH_JWT_ACCESS_EXPIRY` | `auth.access_expiry` |
| `TALKENT_AUTH_JWT_REFRESH_EXPIRY` | `auth.refresh_expiry` |

## 项目结构

```
cmd/server/              - 服务入口
internal/
  config/                - 配置加载与环境变量覆盖
  auth/                  - JWT 认证（Token 生成/验证 + 中间件 + Handler）
  llm/                   - OpenAI 兼容 LLM 客户端
  role/                  - 角色设定 → 目标推荐 → 维度确定
  session/               - 会话生命周期与对话管理
  memory/                - 滑动窗口记忆与摘要压缩
  analysis/              - 多维度分析引擎与 Markdown 报告
  store/                 - SQLite 持久化
  server/                - HTTP 路由、中间件与静态文件托管
  log/                   - slog 日志初始化
web/                     - Vue 3 Web 前端 SPA
android/                 - Kotlin + Compose Android 客户端
  app/src/main/java/com/talkent/app/
    data/
      api/               - Retrofit 接口 + OkHttp 拦截器 + SSE 客户端
      model/             - API DTO
      local/             - Room 数据库 (entity + dao)
      repository/        - AuthRepo + SessionRepo
    ui/
      login/             - 登录页
      setup/             - 角色设定页
      chat/              - 对话页（SSE 流式 + 语音输入）
      report/            - 分析报告页
      settings/          - 设置页（后端地址 + 登出）
      navigation/        - 导航图
      theme/             - Material 3 主题
    util/                - TokenManager, UrlConfig, SpeechRecorder, TtsPlayer
test/e2e/                - E2E 验证场景与 curl 脚本
```

## API

### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/login` | 登录，获取 JWT token |
| POST | `/api/auth/refresh` | 刷新 access token |

### 业务

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/roles/recommend-goals` | Bearer | 根据角色描述推荐训练目标 |
| POST | `/api/roles/recommend-dimensions` | Bearer | 根据角色与目标推荐分析维度 |
| POST | `/api/sessions` | Bearer | 创建对话会话 |
| GET | `/api/sessions/:id` | Bearer | 获取会话详情 |
| POST | `/api/sessions/:id/chat` | Bearer | 发送消息 |
| GET | `/api/sessions/:id/chat/stream` | query token | SSE 流式消息 |
| POST | `/api/sessions/:id/end` | Bearer | 结束会话 |
| POST | `/api/sessions/:id/analyze` | Bearer | 触发分析 |
| GET | `/api/sessions/:id/report` | Bearer | 获取分析报告 |
| GET | `/api/sessions/:id/reports` | Bearer | 获取报告列表 |

> 所有 `/api/*` 路由需要 `Authorization: Bearer <token>` header。
> SSE 流式端点使用 `?token=<token>` query param 传递认证。
> `/health` 和 `/api/auth/*` 无需认证。

## 开发

```bash
# 后端
go build ./...          # 构建
go test ./...           # 运行所有测试
go vet ./...            # 静态检查

# Web 前端
cd web
npm install             # 安装依赖
npm run dev             # 开发服务器（:5173，自动代理 API 到 :8080）
npm run build           # 生产构建

# Android
cd android
export ANDROID_HOME=~/android-sdk
./gradlew assembleDebug     # 编译 Debug APK
./gradlew assembleRelease   # 编译 Release APK
./gradlew test               # 运行单元测试
./gradlew clean              # 清理构建产物

# E2E 验证
./test/e2e/curl/run-all.sh              # 运行全部场景
./test/e2e/curl/scenario-6-auth.sh      # JWT 认证场景
BASE_URL=http://localhost:8080 ./test/e2e/curl/run-all.sh 1  # 运行单个场景
```

## License

MIT
