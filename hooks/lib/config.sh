#!/usr/bin/env bash
# config.sh — load merged IFLy configuration from defaults + global + project + env.
#
# Inputs (env-configurable for tests):
#   IFLY_DEFAULTS  — path to shipped defaults.yaml (required)
#   IFLY_GLOBAL    — path to user global config (optional, may not exist)
#   IFLY_PROJECT   — path to project .ifly.yaml (optional, may not exist)
#   IFLY_MODE      — override mode scalar
#   IFLY_GUARD     — override guard.level scalar
#
# Output to stdout: dotted.key=value lines, merged.
# Lists are unioned (global entries first, then project entries, reindexed).
# Lockdown blocks project from loosening guard.level; warns to stderr.

_IFLY_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=parse_defaults.sh
: "${_IFLY_PARSE:=$_IFLY_LIB_DIR/parse_defaults.sh}"

# Map: strict=3 project=2 open=1 off=0. Higher = stricter.
_ifly_level_rank() {
  case "$1" in
    strict)  echo 3 ;;
    project) echo 2 ;;
    open)    echo 1 ;;
    off)     echo 0 ;;
    *)       echo -1 ;;
  esac
}

# _ifly_parse <path>  — prints parsed lines, empty if path empty/missing.
_ifly_parse() {
  local p="$1"
  [[ -z "$p" || ! -f "$p" ]] && return 0
  bash "$_IFLY_PARSE" < "$p"
}

ifly_load_merged_config() {
  : "${IFLY_DEFAULTS:?IFLY_DEFAULTS must be set}"
  local defaults_lines global_lines project_lines
  defaults_lines="$(_ifly_parse "$IFLY_DEFAULTS")"
  global_lines="$(_ifly_parse "${IFLY_GLOBAL:-}")"
  project_lines="$(_ifly_parse "${IFLY_PROJECT:-}")"

  # Partition scalars vs list items (lines matching key[n]=).
  # Scalars: later source overrides. Lists: union in source order.
  declare -A scalars=()
  declare -A list_acc=()   # key -> newline-joined values

  _absorb() {
    local kind="$1" lines="$2"
    while IFS= read -r line; do
      [[ -z "$line" ]] && continue
      if [[ "$line" =~ ^([^=]+)\[([0-9]+)\]=(.*)$ ]]; then
        local lk="${BASH_REMATCH[1]}" lv="${BASH_REMATCH[3]}"
        # Append list entries only if not already present (dedupe while preserving order)
        local existing="${list_acc[$lk]:-}"
        if ! printf '%s\n' "$existing" | grep -Fx -- "$lv" >/dev/null 2>&1; then
          list_acc["$lk"]+="${lv}"$'\n'
        fi
      elif [[ "$line" =~ ^([^=]+)=(.*)$ ]]; then
        scalars["${BASH_REMATCH[1]}"]="${BASH_REMATCH[2]}"
      fi
    done <<<"$lines"
  }

  # Order: defaults -> global -> project, so later scalars override.
  _absorb defaults "$defaults_lines"

  local g_level="${scalars[guard.level]:-}"
  local lockdown_now="${scalars[guard.lockdown]:-false}"

  _absorb global "$global_lines"
  if [[ -n "${scalars[guard.lockdown]:-}" ]]; then
    lockdown_now="${scalars[guard.lockdown]}"
  fi

  # Hold global guard.level to enforce lockdown check against project.
  g_level="${scalars[guard.level]:-$g_level}"

  # For project, gate guard.level through lockdown check.
  local proj_level_req=""
  local filtered_project=""
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    if [[ "$line" =~ ^guard\.level=(.+)$ ]]; then
      proj_level_req="${BASH_REMATCH[1]}"
      continue
    fi
    filtered_project+="$line"$'\n'
  done <<<"$project_lines"

  _absorb project "$filtered_project"

  # Apply project-requested guard.level with lockdown check.
  if [[ -n "$proj_level_req" ]]; then
    if [[ "$lockdown_now" == "true" ]]; then
      local g_rank p_rank
      g_rank="$(_ifly_level_rank "$g_level")"
      p_rank="$(_ifly_level_rank "$proj_level_req")"
      if [[ "$p_rank" -lt "$g_rank" ]]; then
        echo "ifly: project guard.level '$proj_level_req' ignored (global lockdown enabled)" >&2
      else
        scalars[guard.level]="$proj_level_req"
      fi
    else
      scalars[guard.level]="$proj_level_req"
    fi
  fi

  # Env overrides.
  [[ -n "${IFLY_MODE:-}" ]]  && scalars[mode]="$IFLY_MODE"
  [[ -n "${IFLY_GUARD:-}" ]] && scalars[guard.level]="$IFLY_GUARD"

  # Emit scalars (sorted for determinism).
  local k
  for k in "${!scalars[@]}"; do
    printf '%s=%s\n' "$k" "${scalars[$k]}"
  done | LC_ALL=C sort

  # Emit unioned list entries, reindexed from 0.
  for k in "${!list_acc[@]}"; do
    local idx=0
    while IFS= read -r v; do
      [[ -z "$v" ]] && continue
      printf '%s[%d]=%s\n' "$k" "$idx" "$v"
      idx=$((idx+1))
    done <<<"${list_acc[$k]}"
  done | LC_ALL=C sort
}
