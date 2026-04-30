#!/usr/bin/env bash
# E2E Scenario 3: Empty input boundaries
# Usage: BASE_URL=http://localhost:8080 ./test/e2e/curl/scenario-3-empty-input.sh
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "=== Scenario 3: Empty Input Boundaries ==="
echo "BASE_URL=$BASE_URL"
echo ""

# Test 1: Empty role description
echo "--- Test 3.1: Empty Role Description ---"
RESP=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/roles/recommend-goals" \
  -H "Content-Type: application/json" \
  -d '{"role_description":""}')
HTTP_CODE=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | head -n -1)
echo "HTTP $HTTP_CODE: $BODY"
if [ "$HTTP_CODE" -ge 400 ] && [ "$HTTP_CODE" -lt 500 ]; then
  echo "PASS: got expected 4xx"
else
  echo "FAIL: expected 4xx, got $HTTP_CODE"
  exit 1
fi
echo ""

# Test 2: Empty chat message
echo "--- Test 3.2: Empty Chat Message ---"
# First create a session
SESSION=$(curl -s -X POST "$BASE_URL/api/sessions" \
  -H "Content-Type: application/json" \
  -d '{"role_description":"测试","scenario":"test","role_type":"测试","goals":[{"name":"t","description":"t"}],"dimensions":[{"name":"d","description":"d"}],"round_limit":3}')
SESSION_ID=$(echo "$SESSION" | jq -r '.session_id // empty')
if [ -z "$SESSION_ID" ]; then
  echo "FAIL: could not create session for test"
  exit 1
fi

RESP=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/sessions/$SESSION_ID/chat" \
  -H "Content-Type: application/json" \
  -d '{"content":""}')
HTTP_CODE=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | head -n -1)
echo "HTTP $HTTP_CODE: $BODY"
if [ "$HTTP_CODE" -ge 400 ] && [ "$HTTP_CODE" -lt 500 ]; then
  echo "PASS: got expected 4xx"
else
  echo "FAIL: expected 4xx, got $HTTP_CODE"
  exit 1
fi
echo ""

# Test 3: Nonexistent session report
echo "--- Test 3.3: Nonexistent Session Report ---"
RESP=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/sessions/nonexistent-id/report")
HTTP_CODE=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | head -n -1)
echo "HTTP $HTTP_CODE: $BODY"
if [ "$HTTP_CODE" = "404" ]; then
  echo "PASS: got 404 as expected"
else
  echo "FAIL: expected 404, got $HTTP_CODE"
  exit 1
fi
echo ""

# Test 4: Analyze active session (should get 409)
echo "--- Test 3.4: Analyze Active Session ---"
RESP=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/sessions/$SESSION_ID/analyze")
HTTP_CODE=$(echo "$RESP" | tail -1)
BODY=$(echo "$RESP" | head -n -1)
echo "HTTP $HTTP_CODE: $BODY"
if [ "$HTTP_CODE" = "409" ]; then
  echo "PASS: got 409 conflict as expected"
else
  echo "FAIL: expected 409, got $HTTP_CODE"
  exit 1
fi
echo ""

echo "=== Scenario 3: ALL PASSED ==="
