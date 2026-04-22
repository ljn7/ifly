#!/usr/bin/env bash
# Sanity checks for tests/lib/assert.sh itself.
set -u
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/assert.sh
source "$SCRIPT_DIR/lib/assert.sh"

# This test drives assert_allow/assert_block against a dummy script.
DUMMY="$(mktemp)"
trap 'rm -f "$DUMMY"' EXIT
cat > "$DUMMY" <<'EOF'
#!/usr/bin/env bash
input="$(cat)"
case "$input" in
  *ALLOW*) echo '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}'; exit 0 ;;
  *BLOCK*) echo '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"dummy block reason"}}'; exit 2 ;;
  *)       echo '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}'; exit 0 ;;
esac
EOF
chmod +x "$DUMMY"

# Override guard.sh location for the harness self-test.
GUARD_SCRIPT="$DUMMY"
export GUARD_SCRIPT

test_harness_allow_passes() {
  assert_allow 'ALLOW input'
}

test_harness_block_passes_when_reason_matches() {
  assert_block 'BLOCK input' "dummy block"
}

test_harness_block_fails_when_reason_mismatches() {
  # This test asserts that a wrong reason substring produces failure.
  # We invert: run in a subshell, expect non-zero.
  if (assert_block 'BLOCK input' "nope" 2>/dev/null); then
    return 1
  fi
  _pass
  return 0
}

run_all_tests
