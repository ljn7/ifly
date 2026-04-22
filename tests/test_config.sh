#!/usr/bin/env bash
set -u
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../hooks/lib/config.sh
source "$SCRIPT_DIR/../hooks/lib/config.sh"

pass=0; fail=0
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

_has() {
  if grep -qE "^$2\$" <<<"$1"; then printf '[PASS] %s\n' "$3"; pass=$((pass+1))
  else printf '[FAIL] %s\n  expected line: %s\n  got:\n%s\n' "$3" "$2" "$1" >&2; fail=$((fail+1)); fi
}

# Fake defaults.yaml
DEFAULTS="$TMP/defaults.yaml"
cat > "$DEFAULTS" <<'YAML'
version: 1
mode: minimal
guard:
  level: strict
  lockdown: false
  tools:
    bash: true
    edit: true
  additional_dirs: []
  blocked_commands: []
YAML

# Case 1: defaults only
out="$(IFLY_DEFAULTS="$DEFAULTS" IFLY_GLOBAL="" IFLY_PROJECT="" ifly_load_merged_config)"
_has "$out" 'mode=minimal'            "defaults: mode"
_has "$out" 'guard\.level=strict'     "defaults: guard.level"
_has "$out" 'guard\.tools\.bash=true' "defaults: tools.bash"

# Case 2: global overrides mode, adds to list
GLOBAL="$TMP/global.yaml"
cat > "$GLOBAL" <<'YAML'
mode: verbose
guard:
  additional_dirs:
    - /tmp/shared
YAML
out="$(IFLY_DEFAULTS="$DEFAULTS" IFLY_GLOBAL="$GLOBAL" IFLY_PROJECT="" ifly_load_merged_config)"
_has "$out" 'mode=verbose'                              "global: mode override"
_has "$out" 'guard\.additional_dirs\[0\]=/tmp/shared'   "global: additional_dirs[0]"

# Case 3: project adds to list (union, not replace)
PROJECT="$TMP/project.yaml"
cat > "$PROJECT" <<'YAML'
guard:
  additional_dirs:
    - /tmp/project
YAML
out="$(IFLY_DEFAULTS="$DEFAULTS" IFLY_GLOBAL="$GLOBAL" IFLY_PROJECT="$PROJECT" ifly_load_merged_config)"
_has "$out" 'guard\.additional_dirs\[0\]=/tmp/shared'  "union: global entry kept"
_has "$out" 'guard\.additional_dirs\[1\]=/tmp/project' "union: project entry added"

# Case 4: env overrides everything
out="$(IFLY_DEFAULTS="$DEFAULTS" IFLY_GLOBAL="$GLOBAL" IFLY_PROJECT="$PROJECT" IFLY_GUARD=open IFLY_MODE=silent ifly_load_merged_config)"
_has "$out" 'guard\.level=open'  "env: IFLY_GUARD wins"
_has "$out" 'mode=silent'        "env: IFLY_MODE wins"

# Case 5: lockdown prevents project from loosening
GLOBAL2="$TMP/global2.yaml"
cat > "$GLOBAL2" <<'YAML'
guard:
  level: strict
  lockdown: true
YAML
PROJECT2="$TMP/project2.yaml"
cat > "$PROJECT2" <<'YAML'
guard:
  level: open
YAML
out="$(IFLY_DEFAULTS="$DEFAULTS" IFLY_GLOBAL="$GLOBAL2" IFLY_PROJECT="$PROJECT2" ifly_load_merged_config 2>/dev/null)"
_has "$out" 'guard\.level=strict' "lockdown: project open ignored"

# Case 6: lockdown allows project to tighten
PROJECT3="$TMP/project3.yaml"
cat > "$PROJECT3" <<'YAML'
guard:
  level: off
YAML
GLOBAL3="$TMP/global3.yaml"
cat > "$GLOBAL3" <<'YAML'
guard:
  level: open
  lockdown: true
YAML
# off is looser than open -> ignored
out="$(IFLY_DEFAULTS="$DEFAULTS" IFLY_GLOBAL="$GLOBAL3" IFLY_PROJECT="$PROJECT3" ifly_load_merged_config 2>/dev/null)"
_has "$out" 'guard\.level=open' "lockdown: off looser than open -> ignored"

# Case 7: empty-list defaults aren't treated as entries
out="$(IFLY_DEFAULTS="$DEFAULTS" IFLY_GLOBAL="" IFLY_PROJECT="" ifly_load_merged_config)"
if grep -qE 'additional_dirs\[' <<<"$out"; then
  printf '[FAIL] %s\n  got:\n%s\n' "empty list: no indexed entries" "$out" >&2; fail=$((fail+1))
else
  printf '[PASS] %s\n' "empty list: no indexed entries"; pass=$((pass+1))
fi

printf '\n%d passed, %d failed\n' "$pass" "$fail"
[[ $fail -eq 0 ]]
