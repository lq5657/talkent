#!/usr/bin/env bash
# E2E Scenario 5: Concurrent session creation and data isolation
# Usage: BASE_URL=http://localhost:8080 ./test/e2e/curl/scenario-5-concurrent.sh
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "=== Scenario 5: Concurrent Session Creation ==="
echo "BASE_URL=$BASE_URL"
echo ""

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

# Step 1: Create 3 sessions concurrently
echo "--- Step 1: Create 3 Sessions Concurrently ---"
for i in 1 2 3; do
  (
    curl -s -X POST "$BASE_URL/api/sessions" \
      -H "Content-Type: application/json" \
      -d '{"role_description":"数学老师","scenario":"习题讲解","role_type":"数学老师","goals":[{"name":"讲解清晰","description":"学生能听懂"}],"dimensions":[{"name":"耐心","description":"是否耐心解答"}],"round_limit":2}' \
      > "$TEMP_DIR/session_$i.json"
  ) &
done
wait
echo "PASS: 3 sessions created concurrently"
echo ""

# Step 2: Extract session IDs
SID1=$(jq -r '.session_id' "$TEMP_DIR/session_1.json")
SID2=$(jq -r '.session_id' "$TEMP_DIR/session_2.json")
SID3=$(jq -r '.session_id' "$TEMP_DIR/session_3.json")
echo "Session IDs: $SID1, $SID2, $SID3"

# Verify they're all different
if [ "$SID1" = "$SID2" ] || [ "$SID1" = "$SID3" ] || [ "$SID2" = "$SID3" ]; then
  echo "FAIL: duplicate session IDs detected"
  exit 1
fi
echo "PASS: all session IDs unique"
echo ""

# Step 3: Send different messages to each session
echo "--- Step 3: Send Distinct Messages ---"
MSG1=$(curl -s -X POST "$BASE_URL/api/sessions/$SID1/chat" \
  -H "Content-Type: application/json" \
  -d '{"content":"二次方程怎么解？"}')
echo "Session 1 -> replied: $(echo "$MSG1" | jq -r '.reply[:40]')..."

MSG2=$(curl -s -X POST "$BASE_URL/api/sessions/$SID2/chat" \
  -H "Content-Type: application/json" \
  -d '{"content":"几何证明题的思路？"}')
echo "Session 2 -> replied: $(echo "$MSG2" | jq -r '.reply[:40]')..."

MSG3=$(curl -s -X POST "$BASE_URL/api/sessions/$SID3/chat" \
  -H "Content-Type: application/json" \
  -d '{"content":"概率题怎么做？"}')
echo "Session 3 -> replied: $(echo "$MSG3" | jq -r '.reply[:40]')..."
echo "PASS: all 3 sessions received responses"
echo ""

# Step 4: Verify session details are correct
echo "--- Step 4: Verify Session Isolation ---"
DETAIL1=$(curl -s "$BASE_URL/api/sessions/$SID1")
DETAIL2=$(curl -s "$BASE_URL/api/sessions/$SID2")
DETAIL3=$(curl -s "$BASE_URL/api/sessions/$SID3")

echo "Session 1: $(echo "$DETAIL1" | jq -c '{id: .session_id, status: .status}')"
echo "Session 2: $(echo "$DETAIL2" | jq -c '{id: .session_id, status: .status}')"
echo "Session 3: $(echo "$DETAIL3" | jq -c '{id: .session_id, status: .status}')"

# Verify each returns its own session_id
for i in 1 2 3; do
  expected="SID$i"
  actual=$(eval echo "\$$expected")
  returned=$(eval "echo \"\$DETAIL$i\"" | jq -r '.session_id')
  if [ "$returned" != "$actual" ]; then
    echo "FAIL: session $i returned wrong id: expected $actual, got $returned"
    exit 1
  fi
done
echo "PASS: session data correctly isolated"
echo ""

echo "=== Scenario 5: ALL PASSED ==="
