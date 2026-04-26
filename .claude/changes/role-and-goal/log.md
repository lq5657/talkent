### Log — role-and-goal

#### 2026-04-26 cc-propose

- 创建 spec.md 和 tasks.md
- 变更大小：medium（单推导链路，单验收故事，单验证集群）
- 依赖：scaffold-project(done)、llm-client(done)
- 澄清决策：
  - 目标推荐：模板匹配优先 + LLM 生成兜底（用户选择"两者结合"）
  - 维度确定：查表优先 + LLM 推导兜底 + 用户最终决定（用户补充"优先查表，不满意再 LLM 推导，用户最终决定"）
- 风险：LLM Prompt 设计质量影响推荐/推导效果；模板匹配的关键词覆盖度

#### 2026-04-26 cc-apply

- 基线：21 tests passed, 6 packages
- 分支：`feat/role-and-goal`（从 main 创建，merge feat/llm-client 引入依赖）
- T1: model.go + template.go — Role/TrainingGoal/Dimension 类型 + StructuredExpression 模板 + MatchTemplate + DimensionsForType
- T2: service.go — RecommendGoals（模板匹配→LLM兜底）+ generateGoals（结构化 Prompt 约束 JSON）
- T3: service.go — RecommendDimensions（查表）+ DeriveDimensions（LLM推导），两个独立方法，由 Handler 层决定何时调用 LLM
- T4: handler.go + server.go + main.go — POST /api/roles/recommend-goals + POST /api/roles/recommend-dimensions（mode=derive）；server.New 改为接收 registerRoutes 回调；main.go 将 llmClient 注入 role.Service
- T5: model_test.go(7) + service_test.go(8) + handler_test.go(11) = 26 tests passed，覆盖 V1-V9
- 最终：47 tests passed, 7 packages, go vet 无问题
- Prompt 注入防护：System Prompt 硬编码常量，用户输入仅出现在 User Prompt，V8 测试验证
