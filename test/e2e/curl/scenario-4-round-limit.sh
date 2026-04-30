#!/usr/bin/env bash
# E2E Scenario 4: Boundary round count (round_limit=1)
# Usage: BASE_URL=http://localhost:8080 ./test/e2e/curl/scenario-4-round-limit.sh
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "=== Scenario 4: Boundary Round Limit (round_limit=1) ==="
echo "BASE_URL=$BASE_URL"
echo ""

# Step 1: Create session with round_limit=1
echo "--- Step 1: Create Session (round_limit=1) ---"
SESSION=$(curl -s -X POST "$BASE_URL/api/sessions" \
  -H "Content-Type: application/json" \
  -d '{"role_description":"友善的聊天伙伴","scenario":"日常对话","role_type":"聊天伙伴","goals":[{"name":"活跃气氛","description":"让对话轻松愉快"}],"dimensions":[{"name":"亲和力","description":"是否让人感到舒适"}],"round_limit":1}')
echo "$SESSION" | jq '.'
SESSION_ID=$(echo "$SESSION" | jq -r '.session_id // empty')
if [ -z "$SESSION_ID" ]; then
  echo "FAIL: no session_id"
  exit 1
fi
ROUND_LIMIT=$(echo "$SESSION" | jq -r '.round_limit // 0')
if [ "$ROUND_LIMIT" != "1" ]; then
  echo "FAIL: expected round_limit=1, got $ROUND_LIMIT"
  exit 1
fi
echo "PASS: session created with round_limit=1"
echo ""

# Step 2: Send first message (should be last)
echo "--- Step 2: First (and last) Chat ---"
CHAT=$(curl -s -X POST "$BASE_URL/api/sessions/$SESSION_ID/chat" \
  -H "Content-Type: application/json" \
  -d '{"content":"你好！"}')
echo "$CHAT" | jq '{reply: .reply[:80], round_info: .round_info}'
IS_LAST=$(echo "$CHAT" | jq -r '.round_info.is_last // false')
CURRENT=$(echo "$CHAT" | jq -r '.round_info.current // 0')
if [ "$IS_LAST" != "true" ]; then
  echo "FAIL: expected is_last=true on round 1"
  exit 1
fi
if [ "$CURRENT" != "1" ]; then
  echo "FAIL: expected current=1, got $CURRENT"
  exit 1
fi
echo "PASS: is_last=true, current=1"
echo ""

# Step 3: Try sending beyond limit
echo "--- Step 3: Send Beyond Limit ---"
RESP=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/sessions/$SESSION_ID/chat" \
  -H "Content-Type: application/json" \
  -d '{"content":"再来一轮"}')
HTTP_CODE=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | head -n -1)
echo "HTTP $HTTP_CODE: $BODY"
if [ "$HTTP_CODE" -ge 400 ]; then
  echo "PASS: beyond-limit message correctly rejected"
else
  echo "FAIL: expected 4xx for beyond-limit message, got $HTTP_CODE"
  exit 1
fi
echo ""

echo "=== Scenario 4: ALL PASSED ==="
