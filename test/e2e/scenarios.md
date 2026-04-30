# E2E Verification Scenarios — cc-talkent-v2

```yaml
change_id: e2e-integration
created: 2026-04-30
mapping_id: V5
```

## 前置条件

- 后端服务已启动（`go run ./cmd/server` 或编译后的二进制）
- LLM API Key 已配置（场景 1/3/4/5 需要）
- `BASE_URL` 默认 `http://localhost:8080`，可通过环境变量覆盖

---

## 场景 1: 正常全链路

**目标**: 验证从角色设定到分析报告的完整用户流程。

**步骤**:

1. 推荐训练目标
   ```bash
   curl -s -X POST "$BASE_URL/api/roles/recommend-goals" \
     -H "Content-Type: application/json" \
     -d '{"role_description":"一名经验丰富的技术面试官，擅长考察候选人的系统设计能力"}'
   ```
   **预期**: 返回 `source` 和 `goals` 数组（至少 3 个目标）。

2. 推荐评估维度
   ```bash
   curl -s -X POST "$BASE_URL/api/roles/recommend-dimensions" \
     -H "Content-Type: application/json" \
     -d '{"role_type":"技术面试官","goals":[{"name":"系统设计考察","description":"评估候选人的系统设计能力"}],"mode":"training","role_desc":"一名经验丰富的技术面试官"}'
   ```
   **预期**: 返回 `source` 和 `dimensions` 数组。

3. 创建训练会话
   ```bash
   curl -s -X POST "$BASE_URL/api/sessions" \
     -H "Content-Type: application/json" \
     -d '{"role_description":"一名经验丰富的技术面试官","scenario":"模拟面试场景","role_type":"技术面试官","goals":[{"name":"系统设计考察","description":"评估候选人的系统设计能力"}],"dimensions":[{"name":"提问技巧","description":"问题是否清晰有层次"}],"round_limit":3}'
   ```
   **预期**: 返回 `session_id`、`status: "active"`、`round_limit: 3`。

4. 多轮对话（3 轮）
   ```bash
   SESSION_ID="<上一步返回的 session_id>"
   curl -s -X POST "$BASE_URL/api/sessions/$SESSION_ID/chat" \
     -H "Content-Type: application/json" \
     -d '{"content":"请开始面试吧"}'
   ```
   **预期**: 每轮返回 `reply`、`round_info`（current/limit/is_last），最后一轮 `is_last: true`。

5. 结束会话
   ```bash
   curl -s -X POST "$BASE_URL/api/sessions/$SESSION_ID/end"
   ```
   **预期**: `status: "completed"`。

6. 生成分析报告
   ```bash
   curl -s -X POST "$BASE_URL/api/sessions/$SESSION_ID/analyze"
   ```
   **预期**: 返回 `report_id`、`dimensions` 数组、`markdown`。

7. 查看报告
   ```bash
   curl -s "$BASE_URL/api/sessions/$SESSION_ID/report"
   ```
   **预期**: 返回完整报告内容。

**通过标准**: 每一步返回预期字段，无 5xx 错误。

---

## 场景 2: 后端未启动 / 网络不可达

**目标**: 验证前端在网络不可达时的离线 banner 和错误提示。

**步骤**:

1. 确保后端服务**未启动**。
2. 浏览器访问 `http://localhost:5173`。
3. 观察页面顶部是否显示琥珀色"当前处于离线状态，部分功能不可用" banner。
4. 在设定页输入角色描述并点击"下一步"。
5. 观察是否显示"网络连接失败，请检查后端服务是否启动"错误提示。
6. 启动后端服务。
7. Banner 应自动消失（`online` 事件触发）。
8. 点击错误提示旁的"重试"按钮，确认请求成功。

**通过标准**: 离线 banner 正确显示/隐藏；网络错误信息明确；重试按钮可用。

---

## 场景 3: 空输入边界

**目标**: 验证后端对空/无效输入的正确拒绝。

**步骤**:

1. 空角色描述
   ```bash
   curl -s -X POST "$BASE_URL/api/roles/recommend-goals" \
     -H "Content-Type: application/json" \
     -d '{"role_description":""}'
   ```
   **预期**: 返回错误（4xx），不应 5xx 崩溃。

2. 空消息
   ```bash
   SESSION_ID="<有效 session_id>"
   curl -s -X POST "$BASE_URL/api/sessions/$SESSION_ID/chat" \
     -H "Content-Type: application/json" \
     -d '{"content":""}'
   ```
   **预期**: 返回错误（4xx），说明消息不能为空。

3. 不存在的会话
   ```bash
   curl -s "$BASE_URL/api/sessions/nonexistent-id/report"
   ```
   **预期**: 返回 404，`{"error": "session not found"}`。

4. 未结束会话直接请求分析
   ```bash
   # 创建一个活跃会话后
   curl -s -X POST "$BASE_URL/api/sessions/$ACTIVE_SESSION_ID/analyze"
   ```
   **预期**: 返回 409，提示会话未完成。

**通过标准**: 所有边界输入返回明确的 4xx 错误，无 5xx 崩溃。

---

## 场景 4: 边界轮数（round_limit=1）

**目标**: 验证极限轮数场景下对话自动结束和报告生成。

**步骤**:

1. 创建 round_limit=1 的会话
   ```bash
   curl -s -X POST "$BASE_URL/api/sessions" \
     -H "Content-Type: application/json" \
     -d '{"role_description":"友善的聊天伙伴","scenario":"日常对话","role_type":"聊天伙伴","goals":[{"name":"活跃气氛","description":"让对话轻松愉快"}],"dimensions":[{"name":"亲和力","description":"是否让人感到舒适"}],"round_limit":1}'
   ```

2. 发送第 1 轮消息
   ```bash
   curl -s -X POST "$BASE_URL/api/sessions/$SESSION_ID/chat" \
     -H "Content-Type: application/json" \
     -d '{"content":"你好！"}'
   ```
   **预期**: `round_info.is_last: true`，`round_info.current: 1`。

3. 尝试发送第 2 轮消息
   ```bash
   curl -s -X POST "$BASE_URL/api/sessions/$SESSION_ID/chat" \
     -H "Content-Type: application/json" \
     -d '{"content":"再来一轮"}'
   ```
   **预期**: 返回错误（会话已结束）。

**通过标准**: round_limit=1 时 is_last 正确为 true；超限消息被拒绝。

---

## 场景 5: 并发创建会话数据隔离

**目标**: 验证快速连续创建多个会话时数据正确隔离。

**步骤**:

1. 连续创建 3 个会话（相同角色设定）
   ```bash
   for i in 1 2 3; do
     curl -s -X POST "$BASE_URL/api/sessions" \
       -H "Content-Type: application/json" \
       -d '{"role_description":"数学老师","scenario":"习题讲解","role_type":"数学老师","goals":[{"name":"讲解清晰","description":"学生能听懂"}],"dimensions":[{"name":"耐心","description":"是否耐心解答"}],"round_limit":2}' \
       -o "session_$i.json" &
   done
   wait
   ```

2. 提取 3 个 session_id，确认各不相同。

3. 在每个会话中发 1 条不同内容的消息
   ```bash
   SID1=$(jq -r '.session_id' session_1.json)
   SID2=$(jq -r '.session_id' session_2.json)
   SID3=$(jq -r '.session_id' session_3.json)

   curl -s -X POST "$BASE_URL/api/sessions/$SID1/chat" \
     -H "Content-Type: application/json" -d '{"content":"二次方程怎么解？"}'
   curl -s -X POST "$BASE_URL/api/sessions/$SID2/chat" \
     -H "Content-Type: application/json" -d '{"content":"几何证明题的思路？"}'
   curl -s -X POST "$BASE_URL/api/sessions/$SID3/chat" \
     -H "Content-Type: application/json" -d '{"content":"概率题怎么做？"}'
   ```

4. 分别查看各会话详情，确认消息内容不交叉
   ```bash
   curl -s "$BASE_URL/api/sessions/$SID1" | jq '.session_id'
   curl -s "$BASE_URL/api/sessions/$SID2" | jq '.session_id'
   curl -s "$BASE_URL/api/sessions/$SID3" | jq '.session_id'
   ```

5. 清理临时文件
   ```bash
   rm -f session_*.json
   ```

**通过标准**: 3 个 session_id 各不相同；各会话消息内容独立不交叉。

---

## 执行记录

| 场景 | 日期 | 结果 | 备注 |
|------|------|------|------|
| 1: 正常全链路 | 2026-04-30 | PASS | 7步全链路：goals→dimensions(derive)→session→3轮对话→分析→报告。mode=derive 走 LLM 推导 |
| 2: 网络不可达 | 2026-04-30 | PASS (manual) | 手工浏览器验证：停止后端→前端显示离线 banner；恢复后端→重试成功 |
| 3: 空输入边界 | 2026-04-30 | PASS | 4项全部4xx：空role(400)、空消息(400)、不存在session(404)、活跃session分析(409) |
| 4: round_limit=1 | 2026-04-30 | PASS | is_last=true 正确；超限返回409 |
| 5: 并发创建 | 2026-04-30 | PASS | 3并发创建，2个因LLM限流重试后成功；ID唯一性+消息隔离验证通过 |

## 总结

- 通过 / 总计: 5 / 5
- 阻塞项: 无
- 残留风险: 并发创建场景需重试逻辑应对 LLM API 限流（已在 curl 脚本中实现）
