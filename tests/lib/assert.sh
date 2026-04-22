#!/usr/bin/env bash
# Test harness for IFLy plugin tests.
# Convention: each test function is named test_<group>_<scenario>. run_all_tests
# discovers and runs them, tracking PASS/FAIL and exiting non-zero on any failure.
set -u

# Default guard script location; overridable for harness self-tests.
: "${GUARD_SCRIPT:=${SCRIPT_DIR:-}/../hooks/guard}"

PASS_COUNT=0
FAIL_COUNT=0
CURRENT_TEST=""

_red()   { printf '\033[31m%s\033[0m' "$1"; }
_green() { printf '\033[32m%s\033[0m' "$1"; }

_fail() {
  FAIL_COUNT=$((FAIL_COUNT + 1))
  printf '[%s] %s — %s\n' "$(_red FAIL)" "$CURRENT_TEST" "$1" >&2
}

_pass() {
  PASS_COUNT=$((PASS_COUNT + 1))
  printf '[%s] %s\n' "$(_green PASS)" "$CURRENT_TEST"
}

assert_allow() {
  local input="$1"
  local out rc
  out="$(printf '%s' "$input" | "$GUARD_SCRIPT" 2>/dev/null)"
  rc=$?
  if [[ $rc -ne 0 ]]; then
    _fail "expected allow (exit 0), got exit $rc; stdout=$out"
    return 1
  fi
  if [[ "$out" != *'"permissionDecision"'*'"allow"'* ]]; then
    _fail "expected allow permissionDecision in stdout, got: $out"
    return 1
  fi
  _pass
  return 0
}

assert_block() {
  local input="$1"
  local reason_substr="$2"
  local out rc
  out="$(printf '%s' "$input" | "$GUARD_SCRIPT" 2>/dev/null)"
  rc=$?
  if [[ $rc -ne 2 ]]; then
    _fail "expected block (exit 2), got exit $rc; stdout=$out"
    return 1
  fi
  if [[ "$out" != *"$reason_substr"* ]]; then
    _fail "expected reason substring '$reason_substr' in stdout, got: $out"
    return 1
  fi
  _pass
  return 0
}

run_all_tests() {
  local fn
  while IFS= read -r fn; do
    CURRENT_TEST="$fn"
    "$fn" || true
  done < <(declare -F | awk '$3 ~ /^test_/ {print $3}')

  printf '\n%d passed, %d failed\n' "$PASS_COUNT" "$FAIL_COUNT"
  [[ $FAIL_COUNT -eq 0 ]]
}
