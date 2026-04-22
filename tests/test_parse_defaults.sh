#!/usr/bin/env bash
set -u
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PARSER="$SCRIPT_DIR/../hooks/lib/parse_defaults.sh"

pass=0; fail=0
_check() {
  local name="$1" input="$2" expected="$3"
  local actual
  actual="$(printf '%s\n' "$input" | bash "$PARSER")"
  if [[ "$actual" == "$expected" ]]; then
    printf '[PASS] %s\n' "$name"; pass=$((pass+1))
  else
    printf '[FAIL] %s\n  want: %s\n  got:  %s\n' "$name" "$expected" "$actual" >&2
    fail=$((fail+1))
  fi
}

_check "flat scalars" \
'mode: minimal
version: 1' \
'mode=minimal
version=1'

_check "nested map" \
'guard:
  level: strict
  lockdown: false' \
'guard.level=strict
guard.lockdown=false'

_check "list of strings" \
'guard:
  additional_dirs:
    - /tmp/a
    - /tmp/b' \
'guard.additional_dirs[0]=/tmp/a
guard.additional_dirs[1]=/tmp/b'

_check "quoted string preserved" \
'sensitive:
  - "~/.ssh/"
  - "/etc/shadow"' \
'sensitive[0]=~/.ssh/
sensitive[1]=/etc/shadow'

_check "comments ignored" \
'# top comment
mode: minimal  # inline ignored
# trailing' \
'mode=minimal'

_check "blank lines ignored" \
'

mode: minimal

' \
'mode=minimal'

_check "nested list under nested map" \
'guard:
  tools:
    bash: true
    edit: false' \
'guard.tools.bash=true
guard.tools.edit=false'

_check "deep nesting 3 levels" \
'a:
  b:
    c: v' \
'a.b.c=v'

printf '\n%d passed, %d failed\n' "$pass" "$fail"
[[ $fail -eq 0 ]]
