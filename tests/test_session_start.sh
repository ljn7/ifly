#!/usr/bin/env bash
set -u
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOOK="$SCRIPT_DIR/../hooks/session-start"
export IFLY_DEFAULTS="$SCRIPT_DIR/../defaults.yaml"

pass=0; fail=0
_contains() {
  if [[ "$2" == *"$3"* ]]; then printf '[PASS] %s\n' "$1"; pass=$((pass+1))
  else printf '[FAIL] %s\n  got: %s\n' "$1" "$2" >&2; fail=$((fail+1)); fi
}

# Case 1: no state, no project -> defaults
out="$(IFLY_GLOBAL="" IFLY_PROJECT="" IFLY_STATE="" bash "$HOOK")"
_contains "default mode minimal" "$out" "mode: minimal"
_contains "default guard strict"  "$out" "guard: strict"
_contains "skill reference"       "$out" "ifly-mode-minimal"

# Case 2: state overrides mode
STATE="$(mktemp)"
cat > "$STATE" <<'YAML'
active_mode: silent
YAML
out="$(IFLY_GLOBAL="" IFLY_PROJECT="" IFLY_STATE="$STATE" bash "$HOOK")"
_contains "state active_mode silent" "$out" "mode: silent"
_contains "silent skill ref"          "$out" "ifly-mode-silent"
rm -f "$STATE"

# Case 3: project overrides both
PROJ="$(mktemp)"
cat > "$PROJ" <<'YAML'
mode: verbose
YAML
out="$(IFLY_GLOBAL="" IFLY_PROJECT="$PROJ" IFLY_STATE="" bash "$HOOK")"
_contains "project mode verbose" "$out" "mode: verbose"
rm -f "$PROJ"

printf '\n%d passed, %d failed\n' "$pass" "$fail"
[[ $fail -eq 0 ]]
