#!/usr/bin/env bash
# split_command.sh — split a shell command string into subcommands,
# tracking a virtual PWD across `cd` invocations in the chain.
#
# Emits one line per subcommand: "<pwd><TAB><command>"
# Usage: ifly_split_command <starting_pwd> <command_string>
#
# Supported separators: ; && || | . Subshells, redirections with embedded
# separators, and quoted separators are NOT understood — conservative bailout
# would require a real parser. We treat any unquoted separator as a split.

ifly_split_command() {
  local pwd_now="$1"
  local cmd="$2"
  local -a parts=()
  # _ifly_split_on_ops prints one subcommand per line; collect into an array for portability
  while IFS= read -r line; do
    parts+=("$line")
  done < <(_ifly_split_on_ops "$cmd")

  local sub dir target
  for sub in "${parts[@]}"; do
    # Trim outer whitespace
    sub="${sub#"${sub%%[![:space:]]*}"}"
    sub="${sub%"${sub##*[![:space:]]}"}"
    [[ -z "$sub" ]] && continue

    printf '%s\t%s\n' "$pwd_now" "$sub"

    # If this subcommand is a cd, update virtual pwd for subsequent subcommands.
    if [[ "$sub" =~ ^cd[[:space:]]+([^[:space:]]+) ]]; then
      target="${BASH_REMATCH[1]}"
      # Strip surrounding quotes
      target="${target%\"}"; target="${target#\"}"
      target="${target%\'}"; target="${target#\'}"
      if [[ "$target" == /* ]]; then
        dir="$target"
      else
        dir="$pwd_now/$target"
      fi
      # Normalize via ifly_resolve_path (fallbacks in path_resolve.sh)
      if command -v ifly_resolve_path >/dev/null 2>&1; then
        pwd_now="$(ifly_resolve_path "$dir")"
      else
        if command -v realpath >/dev/null 2>&1 && realpath -m "$dir" >/dev/null 2>&1; then
          pwd_now="$(realpath -m "$dir")"
        else
          pwd_now="$(cd "$dir" 2>/dev/null && pwd)" || pwd_now="$dir"
        fi
      fi
    fi
  done
}

# _ifly_split_on_ops <cmd>
# Prints subcommands split on unquoted ; && || |
_ifly_split_on_ops() {
  local cmd="$1"
  local i c inq='' current='' next=''
  local len=${#cmd}

  for (( i=0; i<len; i++ )); do
    c="${cmd:i:1}"
    next="${cmd:i+1:1}"

    if [[ -n "$inq" ]]; then
      current+="$c"
      if [[ "$c" == "$inq" ]]; then inq=''; fi
      continue
    fi
    if [[ "$c" == "'" || "$c" == '"' ]]; then
      inq="$c"; current+="$c"; continue
    fi
    # Double-char ops
    if [[ "$c$next" == "&&" || "$c$next" == "||" ]]; then
      printf '%s\n' "$current"
      current=''
      i=$((i+1))
      continue
    fi
    if [[ "$c" == ";" || "$c" == "|" ]]; then
      printf '%s\n' "$current"
      current=''
      continue
    fi
    current+="$c"
  done
  printf '%s\n' "$current"
}
