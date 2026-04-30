#!/usr/bin/env bash
# Run all E2E curl scenarios
# Usage: BASE_URL=http://localhost:8080 ./test/e2e/curl/run-all.sh
#   or:   ./test/e2e/curl/run-all.sh [scenario-number]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "========================================="
echo "  cc-talkent-v2 E2E Verification Suite"
echo "  BASE_URL=$BASE_URL"
echo "========================================="
echo ""

run_scenario() {
  local name="$1"
  local file="$2"
  echo ">>> Running: $name"
  if bash "$file"; then
    echo ">>> RESULT: $name — PASSED"
    return 0
  else
    echo ">>> RESULT: $name — FAILED"
    return 1
  fi
}

FAILED=0

if [ "${1:-}" ]; then
  # Run specific scenario
  case "$1" in
    1) run_scenario "Scenario 1: Full Flow" "$SCRIPT_DIR/scenario-1-full-flow.sh" || FAILED=$((FAILED+1)) ;;
    3) run_scenario "Scenario 3: Empty Input" "$SCRIPT_DIR/scenario-3-empty-input.sh" || FAILED=$((FAILED+1)) ;;
    4) run_scenario "Scenario 4: Round Limit" "$SCRIPT_DIR/scenario-4-round-limit.sh" || FAILED=$((FAILED+1)) ;;
    5) run_scenario "Scenario 5: Concurrent" "$SCRIPT_DIR/scenario-5-concurrent.sh" || FAILED=$((FAILED+1)) ;;
    *) echo "Unknown scenario: $1 (available: 1, 3, 4, 5; scenario 2 is manual browser-only)"; exit 1 ;;
  esac
else
  # Run all scenarios (skip scenario 2 — requires manual browser interaction)
  run_scenario "Scenario 1: Full Flow" "$SCRIPT_DIR/scenario-1-full-flow.sh" || FAILED=$((FAILED+1))
  run_scenario "Scenario 3: Empty Input" "$SCRIPT_DIR/scenario-3-empty-input.sh" || FAILED=$((FAILED+1))
  run_scenario "Scenario 4: Round Limit" "$SCRIPT_DIR/scenario-4-round-limit.sh" || FAILED=$((FAILED+1))
  run_scenario "Scenario 5: Concurrent" "$SCRIPT_DIR/scenario-5-concurrent.sh" || FAILED=$((FAILED+1))
fi

echo ""
echo "========================================="
echo "  Results: $FAILED failed"
echo "========================================="

exit $FAILED
