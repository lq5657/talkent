#!/usr/bin/env bash
# E2E Scenario 1: Normal full flow
# Usage: BASE_URL=http://localhost:8080 ./test/e2e/curl/scenario-1-full-flow.sh
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "=== Scenario 1: Normal Full Flow ==="
echo "BASE_URL=$BASE_URL"
echo ""

# Step 1: Recommend goals
echo "--- Step 1: Recommend Goals ---"
GOALS=$(curl -s -X POST "$BASE_URL/api/roles/recommend-goals" \
  -H "Content-Type: application/json" \
  -d '{"role_description":"一名经验丰富的技术面试官，擅长考察候选人的系统设计能力"}')
echo "$GOALS" | jq '.'
GOAL_NAME=$(echo "$GOALS" | jq -r '.goals[0].name // empty')
GOAL_DESC=$(echo "$GOALS" | jq -r '.goals[0].description // empty')
if [ -z "$GOAL_NAME" ]; then
  echo "FAIL: no goals returned"
  exit 1
fi
echo "PASS: goals recommended"
echo ""

# Step 2: Recommend dimensions
echo "--- Step 2: Recommend Dimensions ---"
DIMS=$(curl -s -X POST "$BASE_URL/api/roles/recommend-dimensions" \
  -H "Content-Type: application/json" \
  -d "{\"role_type\":\"技术面试官\",\"goals\":[{\"name\":\"$GOAL_NAME\",\"description\":\"$GOAL_DESC\"}],\"mode\":\"derive\",\"role_desc\":\"一名经验丰富的技术面试官\"}")
echo "$DIMS" | jq '.'
DIM_NAME=$(echo "$DIMS" | jq -r '.dimensions[0].name // empty')
DIM_DESC=$(echo "$DIMS" | jq -r '.dimensions[0].description // empty')
if [ -z "$DIM_NAME" ]; then
  echo "FAIL: no dimensions returned"
  exit 1
fi
echo "PASS: dimensions recommended"
echo ""

# Step 3: Create session
echo "--- Step 3: Create Session ---"
SESSION=$(curl -s -X POST "$BASE_URL/api/sessions" \
  -H "Content-Type: application/json" \
  -d "{\"role_description\":\"一名经验丰富的技术面试官\",\"scenario\":\"模拟面试场景\",\"role_type\":\"技术面试官\",\"goals\":[{\"name\":\"$GOAL_NAME\",\"description\":\"$GOAL_DESC\"}],\"dimensions\":[{\"name\":\"$DIM_NAME\",\"description\":\"$DIM_DESC\"}],\"round_limit\":3}")
echo "$SESSION" | jq '.'
SESSION_ID=$(echo "$SESSION" | jq -r '.session_id // empty')
if [ -z "$SESSION_ID" ]; then
  echo "FAIL: no session_id returned"
  exit 1
fi
echo "PASS: session created: $SESSION_ID"
echo ""

# Step 4: Multi-round chat
for round in 1 2 3; do
  echo "--- Step 4.$round: Chat Round $round ---"
  CHAT=$(curl -s -X POST "$BASE_URL/api/sessions/$SESSION_ID/chat" \
    -H "Content-Type: application/json" \
    -d '{"content":"你好，请开始面试吧"}')
  echo "$CHAT" | jq '{reply: .reply[:50], round_info: .round_info}'
  IS_LAST=$(echo "$CHAT" | jq -r '.round_info.is_last // false')
  if [ "$IS_LAST" = "true" ]; then
    echo "Last round reached at round $round"
    break
  fi
done
echo "PASS: chat rounds completed"
echo ""

# Step 5: End session (may already be ended if is_last was true)
echo "--- Step 5: End Session ---"
END=$(curl -s -X POST "$BASE_URL/api/sessions/$SESSION_ID/end")
echo "$END" | jq '.'
END_STATUS=$(echo "$END" | jq -r '.status // empty')
echo "Session status: $END_STATUS"
echo ""

# Step 6: Analyze
echo "--- Step 6: Analyze ---"
ANALYZE=$(curl -s -X POST "$BASE_URL/api/sessions/$SESSION_ID/analyze")
echo "$ANALYZE" | jq '{report_id: .report_id, model_used: .model_used, dimension_count: (.dimensions | length)}'
REPORT_ID=$(echo "$ANALYZE" | jq -r '.report_id // empty')
if [ -z "$REPORT_ID" ]; then
  echo "FAIL: no report_id returned from analyze"
  exit 1
fi
echo "PASS: analysis complete, report_id=$REPORT_ID"
echo ""

# Step 7: Get report
echo "--- Step 7: Get Report ---"
REPORT=$(curl -s "$BASE_URL/api/sessions/$SESSION_ID/report")
echo "$REPORT" | jq '{report_id: .report_id, model_used: .model_used, dimension_count: (.dimensions | length)}'
echo "PASS: report retrieved"
echo ""

echo "=== Scenario 1: ALL PASSED ==="
