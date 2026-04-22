#!/usr/bin/env bash
# bench_guard.sh — run guard invocations; fail if mean exceeds budget.
set -u
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/../hooks/guard"
CLAUDE_PROJECT_DIR="$(mktemp -d)"
export CLAUDE_PROJECT_DIR
trap 'rm -rf "$CLAUDE_PROJECT_DIR"' EXIT

N="${IFLY_BENCH_N:-100}"
BUDGET_MS="${IFLY_BENCH_BUDGET_MS:-50}"
INPUT='{"tool_name":"Bash","tool_input":{"command":"ls -la"}}'

start_ns=$(date +%s%N)
for ((i=0; i<N; i++)); do
  printf '%s' "$INPUT" | bash "$GUARD" >/dev/null 2>&1
done
end_ns=$(date +%s%N)

total_ms=$(( (end_ns - start_ns) / 1000000 ))
mean_ms=$(( total_ms / N ))

printf 'ran %d invocations in %d ms; mean=%d ms; budget=%d ms\n' "$N" "$total_ms" "$mean_ms" "$BUDGET_MS"

if [[ $mean_ms -gt $BUDGET_MS ]]; then
  printf 'FAIL: mean latency %d ms exceeds %d ms budget\n' "$mean_ms" "$BUDGET_MS" >&2
  exit 1
fi
