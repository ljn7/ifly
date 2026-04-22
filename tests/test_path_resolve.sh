#!/usr/bin/env bash
set -u
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../hooks/lib/path_resolve.sh
source "$SCRIPT_DIR/../hooks/lib/path_resolve.sh"

pass=0; fail=0
_eq() {
  if [[ "$2" == "$3" ]]; then printf '[PASS] %s\n' "$1"; pass=$((pass+1))
  else printf '[FAIL] %s: want=%s got=%s\n' "$1" "$3" "$2" >&2; fail=$((fail+1)); fi
}
_truthy() { if [[ $2 -eq 0 ]]; then printf '[PASS] %s\n' "$1"; pass=$((pass+1))
            else printf '[FAIL] %s (rc=%s)\n' "$1" "$2" >&2; fail=$((fail+1)); fi; }
_falsy()  { if [[ $2 -ne 0 ]]; then printf '[PASS] %s\n' "$1"; pass=$((pass+1))
            else printf '[FAIL] %s (expected nonzero)\n' "$1" >&2; fail=$((fail+1)); fi; }

HOME_BAK="$HOME"
export HOME="/home/tester"

_eq "expand ~"        "$(ifly_resolve_path '~/proj/x')"        "/home/tester/proj/x"
_eq "expand ~ alone"  "$(ifly_resolve_path '~')"               "/home/tester"
_eq "expand \$HOME"   "$(ifly_resolve_path '$HOME/proj')"      "/home/tester/proj"
_eq "expand \${HOME}" "$(ifly_resolve_path '${HOME}/proj')"    "/home/tester/proj"
_eq "collapse .."      "$(ifly_resolve_path '/a/b/../c')"       "/a/c"
_eq "collapse nested .." "$(ifly_resolve_path '/a/b/c/../..')"  "/a"
_eq "absolute passes through" "$(ifly_resolve_path '/etc/passwd')" "/etc/passwd"
_eq "relative to cwd" "$(ifly_resolve_path 'foo/bar' '/tmp/proj')" "/tmp/proj/foo/bar"

# Missing-leaf tolerance (realpath -m)
_eq "missing leaf ok" "$(ifly_resolve_path '/tmp/does-not-exist-xyz/file')" "/tmp/does-not-exist-xyz/file"

# Windows drive-letter normalization. Must produce MSYS form regardless of the
# separator style the caller used. Passing cwd just to prevent relative fallback.
_eq "win: C:/foo"      "$(ifly_resolve_path 'C:/foo' '/tmp')"      "/c/foo"
_eq "win: C:\\foo"      "$(ifly_resolve_path 'C:\foo' '/tmp')"      "/c/foo"
_eq "win: mixed slashes" "$(ifly_resolve_path 'D:\a/b\c' '/tmp')"   "/d/a/b/c"
_eq "win: lowercase drive" "$(ifly_resolve_path 'c:/x' '/tmp')"     "/c/x"
_eq "win: drive alone"  "$(ifly_resolve_path 'C:' '/tmp')"          "/c"
_eq "win: drive-relative C:foo" "$(ifly_resolve_path 'C:foo' '/tmp')" "/c/foo"

export HOME="$HOME_BAK"

# ifly_path_inside
ifly_path_inside "/a/b/c"     "/a/b"    ; _truthy "inside exact"   $?
ifly_path_inside "/a/b"       "/a/b"    ; _truthy "inside self"    $?
ifly_path_inside "/a/bc"      "/a/b"    ; _falsy  "prefix-not-seg" $?
ifly_path_inside "/a"         "/a/b"    ; _falsy  "parent not in"  $?
ifly_path_inside "/etc/passwd" "/home/u" ; _falsy "disjoint"        $?

printf '\n%d passed, %d failed\n' "$pass" "$fail"
[[ $fail -eq 0 ]]
