# Talkent

角色扮演对话训练智能体。

## 快速开始

```bash
# 复制配置文件
cp config.example.yaml config.yaml
# 编辑 config.yaml 填入 LLM API Key
# 启动服务
make run
```

## 项目结构

```
cmd/server/     - 服务入口
internal/
  config/       - 配置管理
  llm/          - LLM 客户端
  role/         - 角色与目标
  session/      - 会话管理
  memory/       - 对话记忆
  analysis/     - 分析引擎
  store/        - 数据持久化
  server/       - HTTP Server
  log/          - 日志
web/            - 前端页面
```
