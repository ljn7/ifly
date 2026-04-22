#!/usr/bin/env bash
# bench_guard.sh — run 100 guard invocations; fail if mean > 50 ms.
set -u
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/../hooks/guard"
export CLAUDE_PROJECT_DIR="$(mktemp -d)"
trap 'rm -rf "$CLAUDE_PROJECT_DIR"' EXIT

N=100
INPUT='{"tool_name":"Bash","tool_input":{"command":"ls -la"}}'

start_ns=$(date +%s%N)
for ((i=0; i<N; i++)); do
  printf '%s' "$INPUT" | bash "$GUARD" >/dev/null 2>&1
done
end_ns=$(date +%s%N)

total_ms=$(( (end_ns - start_ns) / 1000000 ))
mean_ms=$(( total_ms / N ))

printf 'ran %d invocations in %d ms; mean=%d ms\n' "$N" "$total_ms" "$mean_ms"

if [[ $mean_ms -gt 50 ]]; then
  printf 'FAIL: mean latency %d ms exceeds 50 ms budget\n' "$mean_ms" >&2
  exit 1
fi
