# Talkent

角色扮演对话训练智能体。自由设定角色和场景，与 AI 进行 1v1 沉浸式对话训练，结束后获得多维度、结构化的表达分析反馈。

## 技术栈

| 层 | 技术 |
|---|------|
| 后端 | Go 1.24+, net/http, log/slog |
| 前端 | Vue 3 (Composition API), TypeScript, Tailwind CSS v4, Vite |
| 数据库 | SQLite |
| LLM | OpenAI 兼容 API（DeepSeek 等） |

## 快速开始

**前置要求**：Go 1.24+, Node.js 20+

```bash
# 1. 配置
cp config.example.yaml config.yaml
# 编辑 config.yaml，填入 llm.api_key，或通过环境变量注入：
# export TALKENT_LLM_API_KEY="your-api-key"

# 2. 启动后端
make run

# 3. 构建前端（可选，后端已托管静态文件）
cd web && npm install && npm run build && cd ..

# 4. 浏览器访问 http://localhost:8080
```

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

log:
  level: "info"
  file: ""                  # 空字符串 = stdout
```

支持的环境变量覆盖（优先级高于 config.yaml）：

| 环境变量 | 对应配置 |
|----------|----------|
| `TALKENT_LLM_API_KEY` | `llm.api_key` |
| `TALKENT_LLM_BASE_URL` | `llm.base_url` |
| `TALKENT_LLM_MODEL` | `llm.model` |
| `TALKENT_SERVER_PORT` | `server.port` |

## 项目结构

```
cmd/server/          - 服务入口
internal/
  config/            - 配置加载与环境变量覆盖
  llm/               - OpenAI 兼容 LLM 客户端
  role/              - 角色设定 → 目标推荐 → 维度确定
  session/           - 会话生命周期与对话管理
  memory/            - 滑动窗口记忆与摘要压缩
  analysis/          - 多维度分析引擎与 Markdown 报告
  store/             - SQLite 持久化
  server/            - HTTP 路由、中间件与静态文件托管
  log/               - slog 日志初始化
web/                 - Vue 3 前端 SPA
test/e2e/            - E2E 验证场景与 curl 脚本
```

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/roles/recommend-goals` | 根据角色描述推荐训练目标 |
| POST | `/api/roles/recommend-dimensions` | 根据角色与目标推荐分析维度 |
| POST | `/api/sessions` | 创建对话会话 |
| GET | `/api/sessions/:id` | 获取会话详情 |
| POST | `/api/sessions/:id/chat` | 发送消息 |
| POST | `/api/sessions/:id/end` | 结束会话 |
| POST | `/api/sessions/:id/analyze` | 触发分析 |
| GET | `/api/sessions/:id/report` | 获取分析报告 |

## 开发

```bash
# 后端
go build ./...          # 构建
go test ./...           # 运行所有测试
go vet ./...            # 静态检查

# 前端
cd web
npm install             # 安装依赖
npm run dev             # 开发服务器（:5173，自动代理 API 到 :8080）
npm run build           # 生产构建

# E2E 验证
./test/e2e/curl/run-all.sh              # 运行全部场景
BASE_URL=http://localhost:8080 ./test/e2e/curl/run-all.sh 1  # 运行单个场景
```

## License

MIT
