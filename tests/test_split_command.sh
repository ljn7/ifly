#!/usr/bin/env bash
set -u
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../hooks/lib/split_command.sh
source "$SCRIPT_DIR/../hooks/lib/split_command.sh"

pass=0; fail=0
_eq() {
  if [[ "$2" == "$3" ]]; then printf '[PASS] %s\n' "$1"; pass=$((pass+1))
  else printf '[FAIL] %s\n  want:\n%s\n  got:\n%s\n' "$1" "$3" "$2" >&2; fail=$((fail+1)); fi
}

_eq "single command" \
"$(ifly_split_command '/root' 'ls -la')" \
"$(printf '/root\tls -la')"

_eq "semicolon chain" \
"$(ifly_split_command '/root' 'ls ; pwd')" \
"$(printf '/root\tls\n/root\tpwd')"

_eq "&& chain" \
"$(ifly_split_command '/root' 'make && make install')" \
"$(printf '/root\tmake\n/root\tmake install')"

_eq "|| chain" \
"$(ifly_split_command '/root' 'test -f x || touch x')" \
"$(printf '/root\ttest -f x\n/root\ttouch x')"

_eq "pipe preserves cwd" \
"$(ifly_split_command '/root' 'cat a | grep b')" \
"$(printf '/root\tcat a\n/root\tgrep b')"

_eq "cd updates subsequent" \
"$(ifly_split_command '/root' 'cd /tmp && rm x')" \
"$(printf '/root\tcd /tmp\n/tmp\trm x')"

_eq "cd with relative path" \
"$(ifly_split_command '/root/work' 'cd src && ls')" \
"$(printf '/root/work\tcd src\n/root/work/src\tls')"

_eq "cd .. up one" \
"$(ifly_split_command '/a/b' 'cd .. && pwd')" \
"$(printf '/a/b\tcd ..\n/a\tpwd')"

printf '\n%d passed, %d failed\n' "$pass" "$fail"
[[ $fail -eq 0 ]]
